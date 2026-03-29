package constants

const (
	// DateFormat is the date format for query parameters.
	DateFormat = "02-01-2006"

	// Default query parameter values.
	DefaultPageStr  = "1"
	DefaultLimitStr = "20"
	DefaultSortStr  = "asc"
	DefaultPage     = 1
	DefaultLimit    = 20
	SortAsc         = "asc"
	SortDesc        = "desc"

	// Error messages.
	ErrInvalidStartDate = "Invalid start date format"
	ErrInvalidEndDate   = "Invalid end date format"
	ErrInvalidConversationID = "Invalid conversation ID"

	WhatsAppSource = "whatsapp"

	WebsiteSource = "website"
)
