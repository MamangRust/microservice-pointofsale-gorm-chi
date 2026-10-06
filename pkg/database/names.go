package database

// Database names for each bounded context. The POS system runs five separate
// PostgreSQL instances (one per bounded context), each fronted by its own
// PgBouncer. These constants are the single source of truth referenced by the
// env/config and the Kubernetes / Docker Compose manifests.
const (
	IdentityDB = "pos_identity"
	CatalogDB  = "pos_catalog"
	SalesDB    = "pos_sales"
	MerchantDB = "pos_merchant"
	EmailDB    = "pos_email"
)

// Bounded contexts, in the order they must be migrated/seeded (dependencies
// flow identity -> merchant -> catalog -> sales; email is independent).
var BoundedContexts = []string{"identity", "merchant", "catalog", "sales", "email"}

// ContextPrefix maps a bounded context to the env key prefix that configures its
// PostgreSQL instance (e.g. "sales" -> "DB_SALES" reads DB_SALES_HOST/
// DB_SALES_PORT/DB_SALES_NAME/DB_SALES_USERNAME/DB_SALES_PASSWORD).
var ContextPrefix = map[string]string{
	"identity": "DB_IDENTITY",
	"merchant": "DB_MERCHANT",
	"catalog":  "DB_CATALOG",
	"sales":    "DB_SALES",
	"email":    "DB_EMAIL",
}

// ServiceContext maps a service directory name to the bounded context whose
// database it owns. Used by the migrate tool to group migration files, and by
// the seeder to route seed data to the correct instance.
var ServiceContext = map[string]string{
	"auth":        "identity",
	"user":        "identity",
	"role":        "identity",
	"merchant":    "merchant",
	"cashier":     "merchant",
	"category":    "catalog",
	"product":     "catalog",
	"order":       "sales",
	"order_item":  "sales",
	"transaction": "sales",
	"email":       "email",
}
