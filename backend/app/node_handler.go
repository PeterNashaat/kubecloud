package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// ListNodesInput struct holds data required for listing nodes
type ListNodesInput struct {
	Filter *proxyTypes.NodeFilter `json:"filter"`
	Limit  *proxyTypes.Limit      `json:"limit"`
}

// ListNodesHandler requests all nodes from gridproxy
func (h *Handler) ListNodesHandler(c *gin.Context) {
	//TODO: convert this to param
	var request ListNodesInput

	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid filter/limit payload")
		return
	}

	filter := proxyTypes.NodeFilter{}
	if request.Filter != nil {
		filter = *request.Filter
	}

	healthy := true
	rentable := true

	filter.Healthy = &healthy
	filter.Rentable = &rentable

	limit := proxyTypes.DefaultLimit()
	if request.Limit != nil {
		limit = *request.Limit
	}

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Nodes retrieved successfully", gin.H{
		"total": count,
		"nodes": nodes,
	})
}

// ReserveNodeHandler reserves node for user
func (h *Handler) ReserveNodeHandler(c *gin.Context) {
	nodeIDParam := c.Param("node_id")
	if nodeIDParam == "" {
		Error(c, http.StatusBadRequest, "Node ID is required", "")
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		log.Error().Msg("user ID not found in context")
		Error(c, http.StatusBadRequest, "user ID not found in context", "")
		return
	}

	nodeID64, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	nodeID := uint32(nodeID64)

	userID, ok := userIDVal.(int)
	if !ok {
		InternalServerError(c)
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	filter := proxyTypes.NodeFilter{
		NodeID: &nodeID64,
	}

	// validate user has enough balance for reserving node
	nodes, _, err := h.proxyClient.Nodes(c.Request.Context(), filter, proxyTypes.Limit{})
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}
	if len(nodes) == 0 {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "No nodes are available for rent.", "")
		return
	}
	node := nodes[0]

	// Create identity from mnemonic
	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	usdBalance, err := internal.GetUserBalanceUSD(h.substrateClient, user.Mnemonic, user.Debt)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
	}

	//TODO: check price in month constant
	if usdBalance < node.PriceUsd/24/30 || user.Debt > 0 {
		Error(c, http.StatusBadRequest, "You should at lease have enough balance for one hour", "")
		return
	}

	contractID, err := h.substrateClient.CreateRentContract(identity, nodeID, nil)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	err = h.db.CreateUserNode(&models.UserNodes{
		UserID:     userID,
		ContractID: contractID,
		NodeID:     nodeID,
		CreatedAt:  time.Now(),
	})

	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Node is rented successfully", gin.H{
		"contract_id": contractID,
		"node_id":     nodeID,
	})

}

// ListReservedNodeHandler list reserved nodes for user on tfchain
func (h *Handler) ListReservedNodeHandler(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusBadRequest, "User ID is not found in context", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Error().Msg("Invalid user ID or type")
		InternalServerError(c)
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	twinID, err := h.substrateClient.GetTwinByPubKey(identity.PublicKey())
	fmt.Println(twinID)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	twinID64 := uint64(twinID)
	filter := proxyTypes.NodeFilter{
		RentedBy: &twinID64,
	}

	limit := proxyTypes.DefaultLimit()

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Nodes are retrieved successfully", map[string]interface{}{
		"total": count,
		"nodes": nodes,
	})

}

// UnreserveNodeHandler unreserve node for user
func (h *Handler) UnreserveNodeHandler(c *gin.Context) {
	contractIDParam := c.Param("contract_id")
	if contractIDParam == "" {
		Error(c, http.StatusBadRequest, "Contract ID is required", "")
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusBadRequest, "User ID is not found in context", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Error().Msg("Invalid user ID or type")
		InternalServerError(c)
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	contractID64, err := strconv.ParseUint(contractIDParam, 10, 32)
	if err != nil {
		log.Error().Msg("Invalid contract ID or type")
		InternalServerError(c)
		return
	}
	contractID := uint32(contractID64)

	err = h.substrateClient.CancelContract(identity, uint64(contractID))
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}
	Success(c, http.StatusOK, "Node unreserved successfully", nil)

}
