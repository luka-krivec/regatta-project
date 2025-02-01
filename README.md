# Regatta Manager

## Overview
The Regatta Manager is a web application designed to manage regattas, teams, and race results. It provides a RESTful API for creating, updating, and retrieving information about regattas and their associated teams and race results.

## Features
- Create, read, update, and delete regattas
- Manage teams associated with each regatta
- Record and retrieve race results
- Calculate standings and statistics for teams

## Technologies Used
- Go (Golang)
- Gorilla Mux for routing
- SQLite for the database
- JSON for data interchange

## Getting Started

### Prerequisites
- Go (version 1.16 or higher)
- SQLite

### Installation
1. Clone the repository:
   ```bash
   git clone git@github.com:luka-krivec/regatta-project.git
   cd regatta-project
   ```

2. Install the necessary Go packages:
   ```bash
   go mod tidy
   ```

3. Initialize the database:
   Ensure that the database schema is set up correctly. You may need to create the necessary tables in SQLite.

### Running the Application
1.) To start the API server, run:

 ```bash
go run api/regatta.go
   ```

The server will start on `http://localhost:8081`.

2.) To start the web server, run:

 ```bash
go run api/web-regatta.go
   ```

The web page will start on `http://localhost:8080`.

### API Endpoints
- **Regattas**
  - `POST /api/regattas` - Create a new regatta
  - `GET /api/regattas` - Retrieve all regattas
  - `GET /api/regattas/{id}` - Retrieve a specific regatta
  - `PUT /api/regattas/{id}` - Update a specific regatta
  - `DELETE /api/regattas/{id}` - Delete a specific regatta

- **Teams**
  - `GET /api/regattas/{regattaId}/teams` - Retrieve all teams for a regatta
  - `POST /api/regattas/{regattaId}/teams` - Add a new team to a regatta
  - `PUT /api/regattas/{regattaId}/teams/{teamId}` - Update a specific team
  - `DELETE /api/regattas/{regattaId}/teams/{teamId}` - Delete a specific team

- **Race Results**
  - `POST /api/regattas/{regattaId}/results` - Add race results for a regatta
  - `DELETE /api/regattas/{regattaId}/results` - Clear race results for a regatta

- **Standings**
  - `GET /api/regattas/{regattaId}/standings` - Retrieve standings for a regatta

- **Dashboard Stats**
  - `GET /api/dashboard/stats` - Retrieve dashboard statistics

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under GNU General Public License v3

## Acknowledgments
- Thanks to the Go community for their support and resources.
- Special thanks to the contributors who help improve this

