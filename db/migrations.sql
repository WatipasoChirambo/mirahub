-- Drop in correct order to avoid FK conflicts (child tables first)
DROP TABLE IF EXISTS receipt_items CASCADE;
DROP TABLE IF EXISTS receipts CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS quotations CASCADE;
DROP TABLE IF EXISTS sales CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS product_vehicles CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS vehicles CASCADE;
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
    name VARCHAR(255) NOT NULL UNIQUE
);

-- Products
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    item_code VARCHAR(50),
    hold TEXT,
    name VARCHAR(255) NOT NULL,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    supplier_id INT REFERENCES suppliers(id) ON DELETE SET NULL,
    warehouse_id INT REFERENCES warehouses(id) ON DELETE SET NULL,
    stock INT DEFAULT 0,
    price NUMERIC(10,2) DEFAULT 0.00,
    created_by INT REFERENCES users(id) ON DELETE SET NULL,
    image_url TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Many-to-many: products <-> vehicles
CREATE TABLE product_vehicles (
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    vehicle_id INT REFERENCES vehicles(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, vehicle_id)
);

-- Customers
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
    price NUMERIC(10,2),
    total NUMERIC(10,2)
);

-- Sales
CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    customer_id INT REFERENCES customers(id) ON DELETE SET NULL,
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
    total NUMERIC(10,2),
    status VARCHAR(50) DEFAULT 'pending',
    due_date TIMESTAMP,
    pdf_path VARCHAR(500),
    pdf_generated_at TIMESTAMP
);

-- Quotations
CREATE TABLE quotations (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    quote_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    price NUMERIC(10,2),
    status VARCHAR(50) DEFAULT 'draft',
    valid_until TIMESTAMP,
    notes TEXT,
    pdf_path VARCHAR(500),
    pdf_generated_at TIMESTAMP
);

-- Receipts
CREATE TABLE receipts (
    id SERIAL PRIMARY KEY,
    invoice_id INT REFERENCES invoices(id) ON DELETE SET NULL,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    receipt_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    amount NUMERIC(10,2),
    payment_method VARCHAR(50) DEFAULT 'cash',
    reference_no VARCHAR(100),
    notes TEXT,
    customer_id INT REFERENCES customers(id) ON DELETE SET NULL,
    subtotal NUMERIC(10,2) DEFAULT 0,
    tax_rate NUMERIC(5,2) DEFAULT 0,
    tax_amount NUMERIC(10,2) DEFAULT 0,
    discount NUMERIC(10,2) DEFAULT 0,
    total NUMERIC(10,2),
    status VARCHAR(50) DEFAULT 'completed',
    pdf_path VARCHAR(500),
    pdf_generated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Receipt Items
CREATE TABLE receipt_items (
    id SERIAL PRIMARY KEY,
    receipt_id INTEGER NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_name VARCHAR(255) NOT NULL,
    product_code VARCHAR(100),
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    total DECIMAL(10,2) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- File Attachments
CREATE TABLE file_attachments (
    id SERIAL PRIMARY KEY,
    document_type VARCHAR(50) NOT NULL,
    document_id VARCHAR(100) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    uploaded_by INT REFERENCES users(id) ON DELETE SET NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_receipt_items_receipt_id ON receipt_items(receipt_id);
CREATE INDEX IF NOT EXISTS idx_receipts_customer_id ON receipts(customer_id);
CREATE INDEX IF NOT EXISTS idx_receipts_receipt_date ON receipts(receipt_date);
CREATE INDEX IF NOT EXISTS idx_receipts_reference_no ON receipts(reference_no);
CREATE INDEX IF NOT EXISTS idx_invoices_sale_id ON invoices(sale_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_quotations_product_id ON quotations(product_id);
CREATE INDEX IF NOT EXISTS idx_quotations_status ON quotations(status);
CREATE INDEX IF NOT EXISTS idx_sales_product_id ON sales(product_id);
CREATE INDEX IF NOT EXISTS idx_sales_user_id ON sales(user_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_code ON products(code);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
CREATE INDEX IF NOT EXISTS idx_file_attachments_document ON file_attachments(document_type, document_id);

-- Seed Categories
INSERT INTO categories (id, name) VALUES
(1, 'Electronics'),
(2, 'Accessories'),
(3, 'Automotive')
ON CONFLICT (id) DO NOTHING;

-- Seed Suppliers
INSERT INTO suppliers (id, name, contact_info) VALUES
(1, 'MegaTech', '0123456789'),
(2, 'AutoSuppliers Ltd', '013339991')
ON CONFLICT (id) DO NOTHING;

-- Seed Warehouses
INSERT INTO warehouses (id, name, location, capacity, manager, status) VALUES
(1, 'Main Warehouse', 'Blantyre', 1000, 'John Doe', 'active'),
(2, 'Secondary Warehouse', 'Lilongwe', 500, 'Jane Smith', 'active')
ON CONFLICT (id) DO NOTHING;

-- Seed Vehicles
INSERT INTO vehicles (name) VALUES
('Toyota Hilux'),
('Ford Ranger'),
('Nissan Navara'),
('Toyota Prius'),
('Toyota Aqua'),
('Toyota Hiace')
ON CONFLICT (name) DO NOTHING;

-- Seed User (password: admin123 - already hashed)
INSERT INTO users (id, username, email, phone, password_hash, role) VALUES
(1, 'admin', 'admin@mirahub.com', '0990000000',
 '$2a$12$uMl7jYQZ.A4dHqK5bMEwEu6k3Gak8z0N5L8lYEBeo4Qg.UL1rJ9fy',
 'admin')
ON CONFLICT (id) DO NOTHING;

-- Seed Products with online images
INSERT INTO products (code, item_code, name, category_id, supplier_id, warehouse_id, stock, price, created_by, image_url, description) VALUES
('P001', 'ITM001', 'Dell XPS Laptop', 1, 1, 1, 10, 899.99, 1, 'https://images.unsplash.com/photo-1496181133206-80ce9b88a853?w=500', 'High-performance laptop with 16GB RAM, 512GB SSD, Intel Core i7 processor'),
('P002', 'ITM002', 'Logitech MX Master 3', 2, 1, 1, 50, 29.99, 1, 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=500', 'Ergonomic wireless mouse with Bluetooth and USB connectivity'),
('P003', 'ITM003', 'Premium Oil Filter', 3, 2, 2, 100, 15.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'Premium quality oil filter for most vehicle models'),
('P004', 'ITM004', 'Ceramic Brake Pads', 3, 2, 2, 75, 45.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'Ceramic brake pads with superior stopping power, low dust formula'),
('P005', 'ITM005', 'Maintenance-Free Battery', 3, 2, 2, 30, 129.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'Maintenance-free car battery, 60-month warranty, 650 CCA'),
('P006', 'ITM006', '4K HDMI Cable', 2, 1, 1, 200, 12.99, 1, 'https://images.unsplash.com/photo-1583394838336-acd977736f90?w=500', '6ft 4K HDMI cable with high-speed data transfer, supports 4K@60Hz'),
('P007', 'ITM007', '64GB USB Drive', 2, 1, 1, 150, 19.99, 1, 'https://images.unsplash.com/photo-1583394838336-acd977736f90?w=500', 'High-speed USB 3.0 flash drive, 64GB storage'),
('P008', 'ITM008', 'Synthetic Engine Oil', 3, 2, 2, 80, 35.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'Synthetic engine oil, 5W-30, 5 quart'),
('P009', 'ITM009', 'High-Flow Air Filter', 3, 2, 2, 120, 12.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'High-flow air filter for improved engine performance'),
('P010', 'ITM010', '24" LED Monitor', 1, 1, 1, 15, 199.99, 1, 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=500', '24-inch Full HD monitor with IPS display, 75Hz refresh rate'),
('P011', 'ITM011', 'Mechanical Keyboard', 2, 1, 1, 40, 79.99, 1, 'https://images.unsplash.com/photo-1583394838336-acd977736f90?w=500', 'RGB mechanical keyboard with blue switches'),
('P012', 'ITM012', 'Iridium Spark Plugs', 3, 2, 2, 200, 8.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'Iridium spark plugs set of 4'),
('P013', 'ITM013', 'LED Headlights', 3, 2, 2, 60, 49.99, 1, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9?w=500', 'LED headlight bulbs, pair, 6000K bright white light'),
('P014', 'ITM014', 'HD Webcam', 1, 1, 1, 35, 59.99, 1, 'https://images.unsplash.com/photo-1583394838336-acd977736f90?w=500', '1080p HD webcam with built-in microphone'),
('P015', 'ITM015', 'Car Phone Mount', 2, 1, 1, 90, 24.99, 1, 'https://images.unsplash.com/photo-1583394838336-acd977736f90?w=500', 'Universal car phone mount holder, 360° rotation')
ON CONFLICT (code) DO NOTHING;

-- Seed Customers
INSERT INTO customers (name, email, phone, created_by) VALUES
('Walk-in Customer', NULL, NULL, 1),
('John Smith', 'john@example.com', '+265 888 123 456', 1),
('Sarah Johnson', 'sarah@example.com', '+265 999 789 012', 1)
ON CONFLICT (email) DO NOTHING WHERE email IS NOT NULL;

-- Seed product-vehicle relationships
INSERT INTO product_vehicles (product_id, vehicle_id)
SELECT p.id, v.id
FROM products p, vehicles v
WHERE p.code IN ('P003', 'P004', 'P005', 'P008', 'P009', 'P012', 'P013') 
  AND v.name IN ('Toyota Hilux', 'Ford Ranger', 'Nissan Navara', 'Toyota Prius', 'Toyota Aqua', 'Toyota Hiace')
ON CONFLICT DO NOTHING;

-- Update sequences after manual inserts (MOVED TO THE END)
SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories));
SELECT setval('suppliers_id_seq', (SELECT MAX(id) FROM suppliers));
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('warehouses_id_seq', (SELECT MAX(id) FROM warehouses));
SELECT setval('vehicles_id_seq', (SELECT MAX(id) FROM vehicles));
SELECT setval('products_id_seq', (SELECT MAX(id) FROM products));
SELECT setval('customers_id_seq', (SELECT MAX(id) FROM customers));
SELECT setval('receipts_id_seq', (SELECT MAX(id) FROM receipts));