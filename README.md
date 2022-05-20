# Appointment Service

Appointment service is a simple web service written in Golang, backed by SQLite 3, that allows consumers to schedule an appointment with a trainer.

## How do I run this?

A few ways: Docker or Nix. 

### Do you use Nix?

To initialize and run the service with Nix, execute the following:

```bash
nix-shell
make run
```

The service will now be running on port 8080 by default.

### Do you use Docker?

To run the service with Docker, run the following:

```bash
docker build . -t appointment-service
docker run -it --rm -p 8080:8080 appointment-service
```

This will build the Docker container, then execute it, mapping port 8080 to the server's port in the container.

## What's the API look like?

The API has two endpoints: create and get-by-trainer.

### POST /appointment - Creates an appointment

To create an appointment, execute an HTTP POST request to `/appointment` with the following JSON body:

```json
{
  "id": "id",
  "trainer_id": "trainer_id",
  "user_id": "user_id",
  "starts_at": "<RFC3339/ISO 8601 time>",
  "ends_at": "<RFC3339/ISO 8601 time>"
}
```

The `id` field is optional - if not specified, the server will generate a UUID before storing the appointment in the database.

### GET /appointment/trainer/:trainer_id - Get appointments for trainer

To get the appointments for a trainer, execute an HTTP GET request to `/appointment/:trainer_id`, where `:trainer_id` is replaced by a valid trainer ID.
To get the appointments for a trainer within a time frame, perform the same HTTP GET request, but specify the `starts_at` and `ends_at` query parameters: 
```
/appointment/trainer/:trainer_id?starts_at=:start&ends_at=:end
```

where `:start` and `:end` are replaced by either a valid RFC3339/ISO 8601 string, or a Unix millisecond timestamp.
