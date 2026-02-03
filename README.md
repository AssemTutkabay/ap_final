# Beauty Salon Booking System (AP1)

Run:
go run .

Test:
http://localhost:8080/health
# Booking Service API

This is a simple salon booking service API with concurrency support for audit logging.

## Architecture 

- internal/domain — models: Booking, Service, User and etc.
- internal/store/memory — data storage in memory
- internal/service — business logic (BookingService, audit worker)
- internal/http/handlers — HTTP handlers
- internal/app/app.go — application assembly and routes

There is a background worker, auditCh, that writes events to audit.log without blocking HTTP requests.

## Endpoints

- GET /health — service check
- GET /salons — list of salons
- GET /services — list of services (can be filtered ?salonId=...)
- POST /bookings — make a booking
- GET /bookings?clientId=…&masterId=… — list of bookings 
- DELETE /bookings/{id} — cancellation of booking

## Demo commands (curl): 

- Creating booking (201)
- conflict booking (409)
- cancel booking (200)
- check audit.log
