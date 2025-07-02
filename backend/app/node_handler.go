package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// ListNodesResponse holds the response for reserved nodes
type ListNodesResponse struct {
	Total int               `json:"total"`
	Nodes []proxyTypes.Node `json:"nodes"`
}

// @Summary List nodes
// @Description Retrieves a list of nodes from the grid proxy based on the provided filters.
// @ID list-nodes
// @Accept json
// @Produce json
// @Param healthy query bool false "Filter by healthy nodes (default: true)"
// @Param rentable query bool false "Filter by rentable nodes (default: true)"
// @Param limit query int false "Limit the number of nodes returned (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} APIResponse "Nodes are retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid filter parameters"
// @Failure 500 {object} APIResponse "Internal server error"
// ListNodesHandler requests all nodes from gridproxy
func (h *Handler) ListNodesHandler(c *gin.Context) {
	query := c.Request.URL.Query()

	filter := proxyTypes.NodeFilter{}
	err := queryParamsToStruct(query, &filter)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid filter params")
		return
	}

	limit := proxyTypes.DefaultLimit()
	err = queryParamsToStruct(query, &limit)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid limit params")
		return
	}

	// Force Healthy and Rentable to true
	healthy := true
	rentable := true
	filter.Healthy = &healthy
	filter.Rentable = &rentable

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Nodes retrieved successfully", ListNodesResponse{
		Total: count,
		Nodes: nodes,
	})
}

// @Summary Reserve node
// @Description Reserves a node for a user
// @Tags nodes
// @ID reserve-node
// @Accept json
// @Produce json
// @Param node_id path string true "Node ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/{node_id} [post]
// ReserveNodeHandler reserves node for user
func (h *Handler) ReserveNodeHandler(c *gin.Context) {
	nodeIDParam := c.Param("node_id")
	if nodeIDParam == "" {
		Error(c, http.StatusBadRequest, "Node ID is required", "")
		return
	}

	nodeID64, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}
	nodeID := uint32(nodeID64)

	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	filter := proxyTypes.NodeFilter{
		NodeID: &nodeID64,
	}

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
	// validate user has enough balance for reserving node
	usdBalance, err := internal.GetUserBalanceUSD(h.substrateClient, user.Mnemonic, user.Debt)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
	}

	//TODO: check price in month constant
	if usdBalance < node.PriceUsd/24/30 || user.Debt > 0 {
		Error(c, http.StatusBadRequest, "You should at least have enough balance for one hour", "")
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

// @Summary List reserved nodes
// @Description Returns a list of reserved nodes for a user
// @Tags nodes
// @ID list-reserved-nodes
// @Accept json
// @Produce json
// @Success 201 {array} APIResponse
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes [get]
// ListReservedNodeHandler list reserved nodes for user on tfchain
func (h *Handler) ListReservedNodeHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
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

	Success(c, http.StatusCreated, "Nodes are retrieved successfully", ListNodesResponse{
		Total: count,
		Nodes: nodes,
	})

}

// @Summary Unreserve node
// @Description Unreserve a node for a user
// @Tags nodes
// @ID unreserve-node
// @Accept json
// @Produce json
// @Param contract_id path string true "Contract ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/unreserve/{contract_id} [delete]
// UnreserveNodeHandler unreserve node for user
func (h *Handler) UnreserveNodeHandler(c *gin.Context) {
	contractIDParam := c.Param("contract_id")
	if contractIDParam == "" {
		Error(c, http.StatusBadRequest, "Contract ID is required", "")
		return
	}

	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
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
	Success(c, http.StatusAccepted, "Node unreserved successfully", nil)
}

func queryParamsToStruct(query url.Values, result interface{}) error {
	v := reflect.ValueOf(result).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		paramName := field.Tag.Get("schema")
		if paramName == "" {
			paramName = field.Name
		}
		paramName = strings.Split(paramName, ",")[0]

		paramValues, ok := query[paramName]
		if !ok || len(paramValues) == 0 {
			continue
		}

		switch value.Kind() {
		case reflect.Slice:
			elemType := value.Type().Elem()
			slice := reflect.MakeSlice(value.Type(), 0, len(paramValues))
			for _, pv := range paramValues {
				elem := reflect.New(elemType).Elem()
				if err := setValueFromString(elem, pv); err != nil {
					return err
				}
				slice = reflect.Append(slice, elem)
			}
			value.Set(slice)

		case reflect.Ptr:
			ptr := reflect.New(value.Type().Elem())
			if err := setValueFromString(ptr.Elem(), paramValues[0]); err != nil {
				return err
			}
			value.Set(ptr)

		default:
			if err := setValueFromString(value, paramValues[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

func setValueFromString(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int64, reflect.Int32:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint64, reflect.Uint32:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		return fmt.Errorf("unsupported kind: %s", v.Kind())
	}
	return nil
}
