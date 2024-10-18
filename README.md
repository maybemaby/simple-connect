# simple-connect

## Development

Install go-jet cli
```bash
go install github.com/go-jet/jet/v2/cmd/jet@latest
```

## Database

Migrate up
```bash
DB_URL="postgresql://postgres:postgres@localhost:5432/gopg" make migration-up
```

Migrate down
```bash
DB_URL="postgresql://postgres:postgres@localhost:5432/gopg" make migration-down
```

