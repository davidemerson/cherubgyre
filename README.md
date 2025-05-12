
# cherubgyre

**cherubgyre** is an anonymous community defense social network designed to facilitate secure and private interactions among users. The backend is implemented in Go and is containerized for seamless deployment.

## 🌐 Live Deployment

- **API GitHub Repo**: [cherubgyre-api](https://github.com/davidemerson/cherubgyre-api)
- **Docker Image**: [umerfarooq478/cherubgyre](https://hub.docker.com/r/umerfarooq478/cherubgyre)

---

## 🚀 Quick Start with Docker

To quickly deploy the CherubGyre server using Docker:

```bash
docker pull umerfarooq478/cherubgyre
docker run -d -p 80:8080 --name cherubgyre-container umerfarooq478/cherubgyre
```

Verify it:

```bash
curl http://localhost:80/
```

Expected Response:
```
You've reached cherubgyre
```

---

## 🛠️ Development Setup

For local development and testing:

```bash
git clone https://github.com/davidemerson/cherubgyre.git
cd cherubgyre
GOOS=linux GOARCH=amd64 go build -o main main.go
./main
```

---

## 🧪 API Testing

Automated API tests are maintained in a separate repository:

👉 [cherubgyre-api](https://github.com/davidemerson/cherubgyre-api)

That repo contains:

- A `pytest` suite for end-to-end API verification
- Example requests to `/login`, `/duress`, `/invite`, and more
- Setup instructions for running tests with `pytest`

If you've deployed this server locally or remotely, just update the `BASE_URL` at the top of the test file and run:

```bash
pytest test_cherubgyre_api.py
```

---

## 📦 Building and Publishing Docker Images

```bash
docker build --platform linux/amd64 -t yourusername/cherubgyre:latest .
docker run -d -p 80:8080 --name cherubgyre-container yourusername/cherubgyre
docker push yourusername/cherubgyre:latest  # optional
```

---

## 📁 Project Structure

```
cherubgyre/
├── controllers/       # HTTP request handlers
├── dtos/              # Data Transfer Objects
├── repositories/      # Database interaction logic
├── services/          # Business logic
├── main.go            # Application entry point
├── Dockerfile         # Docker configuration
├── go.mod             # Go module file
├── README.md          # Project documentation
└── ...                # Additional config and assets
```

---

## 🤝 Contributing

1. Fork the repository
2. Create your branch (`git checkout -b feature/my-feature`)
3. Commit your changes
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

---

## 📄 License

Licensed under the [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0-standalone.html)

---

## 📞 Contact

Issues and contributions welcome via GitHub at  
👉 [https://github.com/davidemerson/cherubgyre/issues](https://github.com/davidemerson/cherubgyre/issues)
