**README.md**

# Golang E-Commerce Application

This is a simple Golang-based E-Commerce application built using the Gin web framework. The application provides basic functionality for managing products, user authentication, and shopping cart operations.

## Setup

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/ravelinejunior/golang_ecommerce.git
   cd golang_ecommerce
   ```

2. **Install Dependencies:**
   ```bash
   go get -u
   ```

3. **Database Setup:**
   - Ensure you have a running MongoDB instance.
   - Update the database connection details in `database/database.go`.

4. **Environment Variables:**
   - Create a `.env` file in the root directory with the following content:
     ```env
     PORT=8000
     ```

## Usage

Run the application using the following command:

```bash
go run main.go
```

Access the application at [http://localhost:8000](http://localhost:8000) in your web browser.

## Endpoints

- **User Operations:**
  - Register: `POST /register`
  - Login: `POST /login`
  - Logout: `POST /logout`

- **Product Operations:**
  - List Products: `GET /products`
  - Get Product by ID: `GET /products/:id`

- **Shopping Cart Operations:**
  - Add to Cart: `GET /addtocart`
  - Remove Item from Cart: `GET /removeitem`
  - Cart Checkout: `GET /cartcheckout`
  - Instant Buy: `GET /instantbuy`
  - List Cart Items: `GET /listcart`

- **Address Operations:**
  - Add Address: `POST /addaddress`
  - Edit Home Address: `PUT /edithomeaddress`
  - Edit Work Address: `PUT /editworkaddress`
  - Delete Addresses: `GET /deleteaddresses`

## Configuration

- The application uses environment variables for configuration. Ensure the necessary environment variables are set, as mentioned in the Setup section.

## Dependencies

- [Gin](https://github.com/gin-gonic/gin): Web framework for building the HTTP server.
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver): MongoDB driver for Go.

## Contributing

Feel free to contribute by opening issues or submitting pull requests. Follow the established coding style and conventions.

## License

This project is licensed under the [MIT License](LICENSE).
