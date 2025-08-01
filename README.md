# Hotels Data Pipeline

A Go application that fetches hotel data from multiple supplier APIs, merges and cleans the data, and serves it through a RESTful API.

## 🚀 Quick Start

### Prerequisites
- Go 1.18 or higher
- Redis server running locally

### Installation & Run
```bash
# Install dependencies
go mod tidy

# Start Redis (if not running)
redis-server

# Run the application
go run main.go
```

The application will:
- Fetch hotel data from supplier APIs
- Store data in Redis
- Start HTTP server on `localhost:8085`
- Run scheduled updates every 30 seconds

## 🔌 API Endpoints

### Base URL: `http://localhost:8085/api/v1`

### 1. Health Check
```bash
GET /health
```
**Response:**
```json
{"success":true,"data":"Hotel service is running"}
```

### 2. Get Hotel by ID
```bash
GET /hotels/{id}
```
**Example:**
```bash
curl http://localhost:8085/api/v1/hotels/iJhz
```

### 3. Get Hotels by Destination
```bash
GET /hotels/destination/{id}
```
**Example:**
```bash
curl http://localhost:8085/api/v1/hotels/destination/5432
```

### 4. Get Multiple Hotels
```bash
GET /hotels/range?ids=id1,id2,id3
```
**Example:**
```bash
curl "http://localhost:8085/api/v1/hotels/range?ids=iJhz,SjyX,f8c9"
```

## 📊 Response Format

**Success:**
```json
{
  "success": true,
  "data": <response_data>,
  "count": <number_of_items>
}
```

**Error:**
```json
{
  "success": false,
  "error": "Error message"
}
```

## ⚙️ Configuration

Edit `config/test.yaml` to change:
- Supplier URLs
- Redis connection
- HTTP port
- Cron job interval 