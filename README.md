# Backend Engineer Test

## Backend
### How to run BE
1. cd backend
2. go mod tidy and go get
3. setup local env for database at db/database.go on InitDb function based on your local setting
4. go run main.go

### Notes
- Echo were used to handle HTTP router and it's middleware
- Gorilla/Websocket were used to route the HTTP to become socket connections
- Gorm to handle struct DB
- Sync/mutex to handle synchronization

## Frontend
### How to run FE
1. Open the html file on your browser
