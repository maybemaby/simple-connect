# simple-connect

## Database

Migrate up
```bash
DB_URL="postgresql://postgres:postgres@localhost:5432/gopg" make migration-up
```

Migrate down
```bash
DB_URL="postgresql://postgres:postgres@localhost:5432/gopg" make migration-down
```

