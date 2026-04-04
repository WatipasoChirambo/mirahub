-- Drop in correct order to avoid FK conflicts
DROP TABLE IF EXISTS receipts CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS quotations CASCADE;
DROP TABLE IF EXISTS sales CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS product_vehicles CASCADE;
DROP TABLE IF EXISTS vehicles CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS warehouses CASCADE;
DROP TABLE IF EXISTS suppliers CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Categories
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Suppliers
CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_info TEXT
);

-- Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Warehouses
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location TEXT,
    capacity INT,
    manager VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Vehicles (normalized)
CREATE TABLE vehicles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Products
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    item_code VARCHAR(50) UNIQUE NOT NULL,
    hold TEXT,
    name VARCHAR(255) NOT NULL,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    supplier_id INT REFERENCES suppliers(id) ON DELETE SET NULL,
    warehouse_id INT REFERENCES warehouses(id) ON DELETE SET NULL,
    stock INT DEFAULT 0,
    price NUMERIC(10,2) DEFAULT 0.00,
    created_by INT REFERENCES users(id) ON DELETE SET NULL,
    image_url TEXT
);

-- Many-to-many: products <-> vehicles
CREATE TABLE product_vehicles (
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    vehicle_id INT REFERENCES vehicles(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, vehicle_id)
);

CREATE TABLE customers (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE,
  phone TEXT,
  whatsapp TEXT,
  preferences JSONB DEFAULT '{}',
  segment VARCHAR(50) DEFAULT 'new',
  created_by INT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Orders
CREATE TABLE orders (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
  user_id INT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Order Items
CREATE TABLE order_items (
  id SERIAL PRIMARY KEY,
  order_id INT REFERENCES orders(id) ON DELETE CASCADE,
  product_id INT REFERENCES products(id) ON DELETE CASCADE,
  quantity INT NOT NULL,
  price NUMERIC
);

-- Sales
CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    quantity INT NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    sale_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total NUMERIC(10,2)
);

-- Invoices
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    sale_id INT REFERENCES sales(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    invoice_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total NUMERIC(10,2)
);

-- Quotations
CREATE TABLE quotations (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    quote_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    price NUMERIC(10,2)
);

-- Receipts
CREATE TABLE receipts (
    id SERIAL PRIMARY KEY,
    invoice_id INT REFERENCES invoices(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    receipt_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    amount NUMERIC(10,2)
);

-- Seed Categories
INSERT INTO categories (id, name) VALUES
(1, 'Electronics'),
(2, 'Accessories'),
(3, 'Automotive')
ON CONFLICT DO NOTHING;

-- Seed Suppliers
INSERT INTO suppliers (id, name, contact_info) VALUES
(1, 'MegaTech', '0123456789'),
(2, 'AutoSuppliers Ltd', '013339991')
ON CONFLICT DO NOTHING;

-- Seed Warehouses
INSERT INTO warehouses (id, name, location) VALUES
(1, 'Main Warehouse', 'Blantyre'),
(2, 'Secondary Warehouse', 'Lilongwe')
ON CONFLICT DO NOTHING;

-- Seed Vehicles (IMPORTANT for your vehicle_ids)
INSERT INTO vehicles (id, name) VALUES
(1, 'Toyota Hilux'),
(2, 'Ford Ranger'),
(3, 'Nissan Navara')
ON CONFLICT DO NOTHING;

-- Seed User (password already hashed)
INSERT INTO users (id, username, email, phone, password_hash, role) VALUES
(1, 'admin', 'admin@mirahub.com', '0990000000',
'$2a$12$uMl7jYQZ.A4dHqK5bMEwEu6k3Gak8z0N5L8lYEBeo4Qg.UL1rJ9fy',
'admin')
ON CONFLICT DO NOTHING;

-- Indexes
CREATE INDEX idx_sales_product_id ON sales(product_id);
CREATE INDEX idx_sales_user_id ON sales(user_id);
CREATE INDEX idx_products_category_id ON products(category_id);