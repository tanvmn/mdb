### Introduction:
This is a JSON REST service for movies information management.\
Visit this URL to check the service's health: https://demo-tvmn.work.gd/v1/health.

### Service’s REST APIs:
```text
GET /v1/health

GET /v1/movies
POST /v1/movies
GET /v1/movies/{id}
PATCH /v1/movies/{id}
DELETE /v1/movies/{id}

POST /v1/users
PUT /v1/users/activated *disabled due to cloud's vm blocking any SMTP communication*

POST /v1/tokens/authen
```
### Demo with cURL:

#### Check the service’s health:
```text
curl  https://demo-tvmn.work.gd/v1/health

{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
```
\

#### Create new user:
originally, new user won’t be activated, and an activation token will be sent via new user’s email, but right now, the cloud's vm is blocking any email communication via smtp
```text
curl  https://demo-tvmn.work.gd/v1/users --json '{"name":"John Smith","email":"<enter email>","password":"<enter password>"}'

{
  "user": {
    "id": 1,
    "name": "John Smith",
    "email": "<new user email>",
    "activated": true
  }
}
```

#### Create a stateful token for authentication:
```text
curl 'https://demo-tvmn.work.gd/v1/tokens/authen' --json '{"email":"<enter email>","password":"<enter password>"}'

{
  "authenticationToken": {
    "token": "<authentication token>",
    "expiry": "2026-10-24T15:32:14.939831996+07:00"
  }
}
```

\
#### Get movies and limit page size to 2 records:
```text
curl 'https://demo-tvmn.work.gd/v1/movies?pageSize=2' -H "Authorization: Bearer <authentication token>"

{
  "metadata": {
    "currentPage": 1,
    "pageSize": 2,
    "lastPage": 9,
    "totalRecords": 17
  },
  "movies": [
    {
      "id": 1,
      "title": "Toy Story",
      "year": 1995,
      "runtime": "81 mins",
      "genres": [
        "animation",
        "comedy",
        "adventure"
      ],
      "version": 1
    },
    {
      "id": 2,
      "title": "A Bug's Life",
      "year": 1998,
      "runtime": "95 mins",
      "genres": [
        "animation",
        "comedy",
        "quest"
      ],
      "version": 1
    }
  ]
}
```

\
#### Get movie by ID:
```text
curl 'https://demo-tvmn.work.gd/v1/movies/2' -H "Authorization: Bearer <authentication token>"

{
  "movie": {
    "id": 2,
    "title": "A Bug's Life",
    "year": 1998,
    "runtime": "95 mins",
    "genres": [
      "animation",
      "comedy",
      "quest"
    ],
    "version": 1
  }
}
```

\
#### Full-text search movies by title:
```text
curl 'https://demo-tvmn.work.gd/v1/movies?title=toy' -H "Authorization: Bearer <authentication token>"

{
  "metadata": {
    "currentPage": 1,
    "pageSize": 20,
    "lastPage": 1,
    "totalRecords": 5
  },
  "movies": [
    {
      "id": 1,
      "title": "Toy Story",
      "year": 1995,
      "runtime": "81 mins",
      "genres": [
        "animation",
        "comedy",
        "adventure"
      ],
      "version": 1
    },
    {
      "id": 3,
      "title": "Toy Story 2",
      "year": 1999,
      "runtime": "92 mins",
      "genres": [
        "animation",
        "comedy",
        "adventure"
      ],
      "version": 1
    },
    {
      "id": 11,
      "title": "Toy Story 3",
      "year": 2010,
      "runtime": "103 mins",
      "genres": [
        "animation",
        "comedy",
        "adventure"
      ],
      "version": 1
    },
    {
      "id": 16,
      "title": "Toy Story 4",
      "year": 2019,
      "runtime": "100 mins",
      "genres": [
        "animation",
        "comedy",
        "adventure"
      ],
      "version": 1
    },
    {
      "id": 17,
      "title": "Toy Story 5",
      "year": 2026,
      "runtime": "102 mins",
      "genres": [
        "animation",
        "comedy",
        "adventure"
      ],
      "version": 1
    }
  ]
}
```

\
#### Full-text search movies by genres:
```text
curl 'https://demo-tvmn.work.gd/v1/movies?genres=action' -H "Authorization: Bearer <authentication token>"

{
  "metadata": {
    "currentPage": 1,
    "pageSize": 20,
    "lastPage": 1,
    "totalRecords": 2
  },
  "movies": [
    {
      "id": 6,
      "title": "The Incredibles",
      "year": 2004,
      "runtime": "115 mins",
      "genres": [
        "animation",
        "crime",
        "action"
      ],
      "version": 1
    },
    {
      "id": 15,
      "title": "The Incredibles 2",
      "year": 2018,
      "runtime": "118 mins",
      "genres": [
        "animation",
        "crime",
        "action"
      ],
      "version": 1
    }
  ]
}
```

\
#### Partially modify a movie:
*Before modification:*
```text
curl 'https://demo-tvmn.work.gd/v1/movies/2' -H "Authorization: Bearer <authentication token>"

{
  "movie": {
    "id": 2,
    "title": "A Bug's Life",
    "year": 1998,
    "runtime": "95 mins",
    "genres": [
      "animation",
      "comedy",
      "quest"
    ],
    "version": 1
  }
}
```
*After modification:*
```text
curl 'https://demo-tvmn.work.gd/v1/movies/2' -H "Authorization: Bearer <authentication token>" -X PATCH --json '{"runtime":"96 mins"}'

{
  "movie": {
    "id": 2,
    "title": "A Bug's Life",
    "year": 1998,
    "runtime": "96 mins",
    "genres": [
      "animation",
      "comedy",
      "quest"
    ],
    "version": 2
  }
}
```
now "runtime":"96 mins" and "version":2.

\
#### Create a new movie:
```text
curl 'https://demo-tvmn.work.gd/v1/movies' -H "Authorization: Bearer <authentication token>" --json '{"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","family"]}'

{
  "movie": {
    "id": 18,
    "title": "Moana",
    "year": 2016,
    "runtime": "107 mins",
    "genres": [
      "animation",
      "family"
    ],
    "version": 1
  }
}
```

\
#### Delete a movie by ID:
```text
curl 'https://demo-tvmn.work.gd/v1/movies/18' -H "Authorization: Bearer <authentication token>" -X DELETE

{"message":"succeeded"}
```
*If the same request is made again:*
```text
curl 'https://demo-tvmn.work.gd/v1/movies/18' -H "Authorization: Bearer <authentication token>" -X DELETE

{"status":"404 Not Found", "message":"resource not found"}
```

\
#### Optimistic concurrency control to prevent data race:
*8 **concurrent** requests:*
```text
xargs -I % -P8 curl 'https://demo-tvmn.work.gd/v1/movies/2' -H "Authorization: Bearer <authentication token>" -X PATCH --json '{"runtime":"96 mins"}' < <(printf '%s\n' {1..8})

{
  "status": "429 Too Many Requests",
  "message": "exceeded rate limit, retry later"
}
{
  "status": "429 Too Many Requests",
  "message": "exceeded rate limit, retry later"
}
{
  "movie": {
    "id": 2,
    "title": "A Bug's Life",
    "year": 1998,
    "runtime": "96 mins",
    "genres": [
      "animation",
      "comedy",
      "quest"
    ],
    "version": 6
  }
}
{
  "status": "409 Conflict",
  "message": "edit conflict occurred, try again"
}
{
  "status": "429 Too Many Requests",
  "message": "exceeded rate limit, retry later"
}
{
  "status": "409 Conflict",
  "message": "edit conflict occurred, try again"
}
{
  "movie": {
    "id": 2,
    "title": "A Bug's Life",
    "year": 1998,
    "runtime": "96 mins",
    "genres": [
      "animation",
      "comedy",
      "quest"
    ],
    "version": 7
  }
}
{
  "status": "409 Conflict",
  "message": "edit conflict occurred, try again"
}
```
- 3 of 8 requests, that caused data race, are responded with status 409 Conflict
- 3 of 8 requests, that exceeded the client’s rate limit, are responded with status 429 Too Many Requests

\
#### Rate limiting:
*8 **concurrent** requests:*
```text
xargs -I % -P8 curl 'https://demo-tvmn.work.gd/v1/health' -H "Authorization: Bearer <authentication token>" < <(printf '%s\n\n' {1..8})

{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
{
  "status": "429 Too Many Requests",
  "message": "exceeded rate limit, retry later"
}
{
  "environment": "production",
  "status": "available",
  "version": "v1.0.1-0.20261023125932-de1be54d6548"
}
{
  "status": "429 Too Many Requests",
  "message": "exceeded rate limit, retry later"
}
```
- 2 of 8 requests, that exceeded their client’s rate limit, are responded with status 429 Too Many Requests

\
Service actually limits a user’s access by the user's permission(s) (read, write) on movies, but for now all users have both permissions.

\
Instructions to build this project will be added later.
