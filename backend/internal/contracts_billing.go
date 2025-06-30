package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
)

// query MyQuery {
//   contractBillReports(where: {contractID_eq: "1231758", timestamp_lte: ""}) {
//     contractID
//     timestamp
//     amountBilled
//   }
// }

type ContractBillReports struct {
	Reports []Report `json:"reports"`
}

type Report struct {
	ContractID   uint64 `json:"contractID"`
	Timestamp    uint64 `json:"timestamp"`
	AmountBilled uint64 `json:"amountBilled"`
}

// ListContractBillReportsPerMonth returns bill reports for contract ID month ago
func ListContractBillReportsPerMonth(graphqlClient *graphql.GraphQl, contractID uint64, currentTime time.Time) (ContractBillReports, error) {
	monthAgo := currentTime.AddDate(0, -1, 0)

	options := fmt.Sprintf(`(where: {contractID_eq: %v, timestamp_lte: %v, timestamp_gte: %v})`, contractID, currentTime.Unix(), monthAgo.Unix())
	billingReportsCount, err := graphqlClient.GetItemTotalCount("billingReports", options)
	if err != nil {
		return ContractBillReports{}, err
	}
	billingReportsData, err := graphqlClient.Query(fmt.Sprintf(`query MyQuery($billingReportsCount: Int!){
            contractBillReports(where: {contractID_eq: %v, timestamp_lte: %v, timestamp_gte: %v}, limit: $billingReportsCount) {
              contractID
              timestamp
              amountBilled
            }
          }`, contractID, currentTime.Unix(), monthAgo.Unix()),
		map[string]interface{}{
			"billingReportsCount": billingReportsCount,
		})
	if err != nil {
		return ContractBillReports{}, err
	}

	billingReports, err := json.Marshal(billingReportsData)
	if err != nil {
		return ContractBillReports{}, err
	}

	var reports ContractBillReports
	err = json.Unmarshal(billingReports, &reports)
	if err != nil {
		return ContractBillReports{}, err
	}

	return reports, nil
}
