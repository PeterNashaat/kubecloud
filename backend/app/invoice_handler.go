package app

import (
	"fmt"
	"kubecloud/internal"
	"net/http"
	"strconv"

	// "kubecloud/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

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

// ListUserInvoicesHandler lists user invoices by its ID
func (h *Handler) ListUserInvoicesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	_, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

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

func (h *Handler) MonthlyInvoicesHandler(c *gin.Context) {
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
			records, err := h.db.ListUserNodes(user.ID)

			if err != nil {
				log.Error().Err(err).Send()
			}
			fmt.Println(records)
			// h.db.CreateInvoice(&models.Invoice{
			// 	UserID: user.ID,
			// 	//TODO:
			// 	Total: 0,
			// })

		}
	}

}

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
		Error(c, http.StatusNotFound, "Invalid invoice ID", "")
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID)))
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", invoice.FileData)

}
