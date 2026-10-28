# SHUX API
<a href="https://discord.gg/6y7Fh8x">
  <img src="https://discordapp.com/api/guilds/392414185633611776/widget.png?style=shield" alt="Shux Team Discord Server">
</a>

REST API for the [ShuxTeam](https://www.youtube.com/channel/UCt7GNv0mKwyu3SzltispROw)'s Discord Bot.

# Authentication
Almost all REST API endpoints require authentication. You can authenticate your request by sending a token in the
`Authorization` header of your request. Replace `<YOUR-TOKEN>` with the token you get at login:

```
Authorization: Bearer <YOUR-TOKEN>
```

# Endpoints

`Base URL: https://shux.adrephos.com/api/v1`

### Auth Endpoints

| Endpoint      | HTTP Methods | Description                                         |
|---------------|--------------|-----------------------------------------------------|
| auth/refresh  | POST         | Post refresh token to generate new access token     |
| auth/register | POST         | Register new admin                                  |
| auth/login    | POST         | Authenticate and generate access and refresh tokens |

_To use the register endpoint a token given to administrators is needed._

### Server Endpoints

| Endpoint                        | HTTP Methods | Description                            |
|---------------------------------|--------------|----------------------------------------|
| servers                         | GET          | List of IDs of currently added servers |
| servers/{server-id}/leaderboard | GET          | Top five users of a server             |

### User Endpoints

| Endpoint                                 | HTTP Methods                  | Description                     |
|------------------------------------------|-------------------------------|---------------------------------|
| servers/{server_id}/users/{user_id}      | GET, DELETE, PATCH, PUT, POST | Manage user information         |
| servers/{server_id}/users/{user_id}/rank | GET                           | Get the rank of a specific user |

### Role Endpoints

| Endpoint                            | HTTP Methods                  | Description             |
|-------------------------------------|-------------------------------|-------------------------|
| servers/{server_id}/roles           | GET                           | Get all roles           |
| servers/{server_id}/roles/{role_id} | GET, DELETE, PATCH, PUT, POST | Manage role information |

### Channel Endpoints

| Endpoint                                  | HTTP Methods                  | Description                |
|-------------------------------------------|-------------------------------|----------------------------|
| servers/{server_id}/channels              | GET                           | Get all channels           |
| servers/{server_id}/channels/{channel_id} | GET, DELETE, PATCH, PUT, POST | Manage channel information |

# Response format

Responses will have two fields out of three possible:
- success: true if the request was satisfied
- data: the requested resource (sent only when success is true)
- error: provides an error description (sent only when success is false)

# Login

At login you'll get two tokens. The `accessToken` is required to access all endpoints except the auth endpoints, 
this token will expire after ___15 min___. The `refreshToken` should be sent at the `auth/refresh` endpoint body 
to get new access and refresh tokens, this token will expire after ___20 min___.

`request`
```bash
curl --location 'https://shux.adrephos.com/api/v1/auth/login' \
--header 'Content-Type: application/json' \
--data '{
    "username": <username>,
    "password": <password>
}'
```
`response`
```json
{
    "data": {
        "accessToken": "xxxxx.yyyyy.zzzzz",
        "refreshToken": "xxxxx.yyyyy.zzzzz"
    },
    "success": true
}
```

# Running the API

> [!IMPORTANT]  
> To run the API you'll need a running PostgreSQL instance. The schema is created automatically on startup.

### Environment variables

In order for the API to work properly, these enviroment variables need to be set in your system:

| Key             | Description                                              |
|-----------------|-----------------------------------------------------------|
| ACCESS_SECRET   | Secret to sign access tokens                              |
| REFRESH_SECRET  | Secret to sign refresh tokens                              |
| REGISTER_SECRET | Secret to allow registration                               |
| DATABASE_URL    | Postgres connection string (e.g. postgres://user:pass@host:5432/shux) |
| API_PORT        | Port in which the API will run                             |

### Local run

> [!NOTE]  
> To run the API in your machine you need to have [Go installed](https://go.dev/doc/install).

After the enviroment variables are set, you can clone the repository and install the dependencies:

```
$ git clone https://github.com/shuxbot/shux-api
$ cd shux-api
$ go mod tidy
```

Now you can run the API with:
```
$ go run main.go
```
Make an executable:
```
$ go build
```
Or with a custom name:
```
$ go build -o <custom_name>
```
### Docker run

> [!NOTE]  
> Remember to [install Docker](https://go.dev/doc/install) if you haven't already.

Clone the repository:
```
$ git clone https://github.com/shuxbot/shux-api
$ cd shux-api
```

Now you can build the Docker image:
```
$ docker buildx build -t <image_name> .
```

And run the docker, pointing it at your Postgres instance:
```
$ docker run -e ACCES_SECRET="secret" \
-e REFRESH_SECRET="secret" \
-e REGISTER_SECRET="secret" \
-e DATABASE_URL="postgres://user:pass@host:5432/shux" \
-p 3000:3000 --name <container_name> <image_name>
```
*You can see other ways to set [enviroment variables](https://docs.docker.com/engine/reference/commandline/run/#env).*

## Examples

**User information can be requested like this:**

`request`
```bash
curl --location 'https://shux.adrephos.com/api/v1/servers/719567919545319535/users/361159575036231691' \
--header 'Authorization: Bearer <YOUR-TOKEN>'
```
`response`
```json
{
    "data": {
        "id": "361159575036231691",
        "description": "La vida es bella",
        "points": 600.9622791840214
    },
    "success": true
}
```

**To edit a user field, you can use the `PATCH` method:**

`request`
```bash
curl --location --request PATCH 'https://shux.adrephos.com/api/v1/servers/719567919545319535/users/361159575036231691' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <YOUR-TOKEN>' \
--data '{
    "points": 100.2132131
}'
```
`response`
```json
{
    "data": {
        "points": 100.2132131
    },
    "success": true
}
```
**Getting information of a user that does not exist:**

`request`
```bash
curl --location 'https://shux.adrephos.com/api/v1/servers/719567919545319535/users/13241324123412341' \
--header 'Authorization: Bearer <YOUR-TOKEN>'
```
`response`
```json
{
    "error": "Not found",
    "success": false
}
```
