-- Categories
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Suppliers
CREATE TABLE IF NOT EXISTS suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_info TEXT
);

-- Users table with email and phone
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) DEFAULT 'user', -- e.g., admin, user
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Warehouses
CREATE TABLE IF NOT EXISTS warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location TEXT
);

-- Products
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    category_id INT REFERENCES categories(id),
    supplier_id INT REFERENCES suppliers(id),
    warehouse_id INT REFERENCES warehouses(id),
    stock INT DEFAULT 0,
    price NUMERIC(10,2) DEFAULT 0.00,  -- unit price added
    created_by INT REFERENCES users(id)
);

-- Sales with price
CREATE TABLE IF NOT EXISTS sales (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id),
    user_id INT REFERENCES users(id),      -- who made the sale
    quantity INT NOT NULL,
    price NUMERIC(10,2) NOT NULL,         -- price at time of sale
    sale_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Invoices
CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    sale_id INT REFERENCES sales(id),
    user_id INT REFERENCES users(id), -- who issued the invoice
    invoice_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total NUMERIC(10,2)
);

-- Quotations
CREATE TABLE IF NOT EXISTS quotations (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id),
    user_id INT REFERENCES users(id), -- who created the quote
    quote_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    price NUMERIC(10,2)
);

-- Receipts
CREATE TABLE IF NOT EXISTS receipts (
    id SERIAL PRIMARY KEY,
    invoice_id INT REFERENCES invoices(id),
    user_id INT REFERENCES users(id), -- who received the payment
    receipt_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    amount NUMERIC(10,2)
);