package components

const ContractsViewType ViewType = "contracts-view"

// ContractInfo represents a single contract/package entry
type ContractInfo struct {
	Path    string // Full package path
	Name    string // Package name (last part of path)
	Type    string // "realm" or "pure"
	URL     string // URL to navigate to
	IsRealm bool
}

// ContractsData holds data for the contracts list view
type ContractsData struct {
	Title        string
	Contracts    []ContractInfo
	TotalCount   int
	RealmCount   int
	PackageCount int
	Mode         ViewMode
	// Pagination fields
	CurrentPage int
	HasNextPage bool
	HasPrevPage bool
	BaseURL     string
}

// ContractsView creates a new contracts list view component
func ContractsView(data ContractsData) *View {
	return NewTemplateView(ContractsViewType, "renderContracts", data)
}
