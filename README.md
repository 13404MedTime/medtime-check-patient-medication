# CardioApp - Check Patient Medication

This project calculates the best distance time for cardio exercises, such as running or cycling. It helps you analyze and determine your optimal performance based on time and distance data.

## Features

- Calculate best distance over time
- CLI-based interface
- Easy to extend for additional metrics

## Prerequisites

- [Go](https://golang.org/dl/) version 1.21 or higher

## Getting Started

1. Clone the repository:

   ```bash
   cd medtime-check-patient-medication
   

2. Install dependencies:

   
   go mod tidy
   

3. Run the application:

   
   go run cmd/main.go
   

## Project Structure


medtime-check-patient-medication/
├── cmd/
│   └── main.go         # Entry point of the application
├── go.mod              # Go module file
├── go.sum              # Go module checksums
├── handler.go          # Request handler logic
├── request.json        # Sample input JSON
└── README.md           # Project documentation


## License

This project is licensed under the MIT License.
`