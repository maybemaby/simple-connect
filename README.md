# simple-connect

## Database

Migrate up
```bash
DB_URL="postgresql://postgres:postgres@localhost:5432/connectpg" make migration-up
```

Migrate down
```bash
DB_URL="postgresql://postgres:postgres@localhost:5432/connectpg" make migration-down
```

