# LivePoll

LivePoll is a real-time polling web application where users can create polls, share a poll link, and allow an audience to vote. The results are updated live without requiring a page refresh.

## Features

- User Signup and Login
- Authentication for poll creation
- Create polls with multiple options
- Shareable poll links
- Audience voting
- Live vote count updates
- Real-time results without page refresh
- MongoDB data storage
- Redis-based real-time updates
- Responsive and simple user interface

## Tech Stack

### Frontend
- React
- React Router
- JavaScript
- CSS

### Backend
- Go
- Gin Framework
- WebSockets
- JWT Authentication
- bcrypt password hashing

### Database
- MongoDB

### Realtime
- Redis
- Redis Pub/Sub

## Project Structure

```text
LivePoll/
│
├── frontend/
│   └── React application
│
├── backend/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── main.go
│   └── go.mod
│
└── README.md