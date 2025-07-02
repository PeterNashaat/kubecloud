package app

import (
	"errors"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// @Summary Get all invoices
// @Description Returns a list of all invoices
// @ID get-all-invoices
// @Accept json
// @Produce json
// @Success 200 {array} models.Invoice
// @Failure 500 {object} APIResponse
// @Router /invoices [get]
// ListAllInvoicesHandler lists all invoices in system
func (h *Handler) ListAllInvoicesHandler(c *gin.Context) {
	invoices, err := h.db.ListInvoices()
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Invoices are retrieved successfully", map[string]interface{}{
		"invoices": invoices,
	})
}

// @Summary Get invoices
// @Description Returns a list of invoices for a user
// @ID get-invoices
// @Accept json
// @Produce json
// @Success 200 {array} models.Invoice
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/invoice [get]
// ListUserInvoicesHandler lists user invoices by its ID
func (h *Handler) ListUserInvoicesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	invoices, err := h.db.ListUserInvoices(userID)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Invoices are retrieved successfully", map[string]interface{}{
		"invoices": invoices,
	})
}

func (h *Handler) MonthlyInvoicesHandler() {
	for {
		now := time.Now()
		monthLastDay := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
		if now.Day() != monthLastDay.Day() {
			// sleep till last day of month
			time.Sleep(monthLastDay.Sub(now))
		}

		users, err := h.db.ListAllUsers()
		if err != nil {
			log.Error().Err(err).Send()
		}
		for _, user := range users {
			if err = h.createUserInvoice(user, monthLastDay); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}

}

// @Summary Download invoice
// @Description Downloads an invoice by ID
// @ID download-invoice
// @Accept json
// @Produce octet-stream
// @Param invoice_id path string true "Invoice ID"
// @Success 200 {file} Invoice
// @Failure 404 {object} APIResponse "Invoice is not found"
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/invoice/{invoice_id} [get]
func (h *Handler) DownloadInvoiceHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	invoiceID := c.Param("invoice_id")
	if invoiceID == "" {
		Error(c, http.StatusBadRequest, "Invoice ID is required", "")
		return
	}

	id, err := strconv.Atoi(invoiceID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid invoice ID", err.Error())
		return
	}

	invoice, err := h.db.GetInvoice(id)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "Invoice is not found", "")
		return
	}

	// Creating pdf for invoice if it doesn't have it
	if len(invoice.FileData) == 0 {
		user, err := h.db.GetUserByID(userID)
		if err != nil {
			log.Error().Err(err).Send()
			InternalServerError(c)
			return
		}

		pdfContent, err := internal.CreateInvoicePDF(invoice, user)
		if err != nil {
			log.Error().Err(err).Send()
			InternalServerError(c)
			return
		}

		invoice.FileData = pdfContent
		if err := h.db.UpdateInvoicePDF(id, invoice.FileData); err != nil {
			log.Error().Err(err).Send()
			InternalServerError(c)
			return
		}
	}

	if userID != invoice.UserID {
		Error(c, http.StatusNotFound, "User is not authorized to download this invoice", "")
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID)))
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", invoice.FileData)

}

func (h *Handler) createUserInvoice(user models.User, monthLastDay time.Time) error {
	records, err := h.db.ListUserNodes(user.ID)
	now := time.Now()

	if err != nil {
		return err
	}
	var nodeItems []models.NodeItem
	var totalInvoiceCost float64

	for _, record := range records {
		billReports, err := internal.ListContractBillReportsPerMonth(h.graphqlClient, record.ContractID, now)
		if err != nil {
			return err
		}

		totalAmountTFT, err := internal.AmountBilledPerMonth(billReports)
		if err != nil {
			return err
		}
		totalAmountUSD, err := internal.FromTFTtoUSD(h.substrateClient, totalAmountTFT)
		if err != nil {
			return err
		}
		rentRecordStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if record.CreatedAt.After(rentRecordStart) {
			rentRecordStart = record.CreatedAt
		}

		var totalHours int
		cancellationDate, err := internal.GetRentContractCancellationDate(h.firesquidClient, record.ContractID)

		if errors.Is(err, internal.ErrorEventsNotFound) {
			totalHours = GetHoursOfGivenPeriod(rentRecordStart, time.Now())
		} else if err != nil {
			return err
		} else {
			totalHours = GetHoursOfGivenPeriod(rentRecordStart, cancellationDate)
		}

		nodeItems = append(nodeItems, models.NodeItem{
			NodeID:        record.NodeID,
			ContractID:    record.ContractID,
			RentCreatedAt: rentRecordStart,
			PeriodInHours: float64(totalHours),
			Cost:          totalAmountUSD,
		})
		totalInvoiceCost += totalAmountUSD

	}
	invoice := models.Invoice{
		UserID:    user.ID,
		Total:     totalInvoiceCost,
		Nodes:     nodeItems,
		Tax:       0, //TODO:
		CreatedAt: time.Now(),
	}
	file, err := internal.CreateInvoicePDF(invoice, user)
	if err != nil {
		return err
	}
	invoice.FileData = file
	if err = h.db.CreateInvoice(&invoice); err != nil {
		return err
	}

	subject, body := h.mailService.InvoiceMailContent(totalInvoiceCost, h.config.Currency, invoice.ID)
	err = h.mailService.SendMail(h.config.MailSender.Email, user.Email, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func GetHoursOfGivenPeriod(startDate, endDate time.Time) int {
	// Calculate the duration between the first day of the month and the specific date
	duration := endDate.Sub(startDate)
	// Convert the duration to hours
	hours := int(duration.Hours())
	return hours
}
