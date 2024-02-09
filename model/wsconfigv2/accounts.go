package wsconfigv2

type Account struct {
	Name          string                 `json:"name"`
	Role          string                 `json:"role"`
	Options       map[string]interface{} `json:"options"`
	Secret        map[string]interface{} `json:"secret"`
	Metadata      map[string]interface{} `json:"metadata"`
	UserID        string                 `json:"userId"`
	SecretVersion int                    `json:"secretVersion"`
	Category      AccountCategory        `json:"rudderCategory"`
}

type AccountCategory string

const (
	AccountCategorySource        AccountCategory = "source"
	AccountCategoryDestination   AccountCategory = "destination"
	AccountCategoryDataRetention AccountCategory = "dataRetention"
	AccountCategoryWHT           AccountCategory = "wht"
)
