# SPLIT-EXPRESS-BACKEND

**Empowering Seamless Collaboration and Effortless Travel Expense Management**  
Built with the power of **Go**, **MongoDB**, and **JWT**.

![Go](https://img.shields.io/badge/Go-1.21-blue)
![MongoDB](https://img.shields.io/badge/MongoDB-6.x-green)
![License](https://img.shields.io/badge/License-MIT-yellow)

---

## 🚀 Overview

**Split-Express-Backend** is a scalable backend framework tailored for **group travel** and **expense tracking** applications. It is built on a **modular architecture** with focus on:

- ⚡️ Rapid Development
- 🔐 Robust Security (JWT & role-based access control)
- 🧩 Easy Maintainability & Extensibility
- 📦 Clean API Structure for integration

With Go's high concurrency and MongoDB's flexible data modeling, it ensures efficient handling of trips, expenses, and user collaboration.

---

## 📚 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Getting Started](#-getting-started)
- [Usage](#-usage)
- [Testing](#-testing)

---

## ✨ Features

✅ **Authentication & Security**  
- JWT-based authentication  
- Role-based access control for user/group authorization

✅ **Modular Clean Architecture**  
- Clear separation of models, controllers, routes, and helpers  
- Promotes better code organization and testability

✅ **MongoDB Integration**  
- Flexible document modeling for trips, users, and expenses  
- Efficient aggregation queries

✅ **RESTful API Endpoints**  
- Full support for trip creation, member management, and expense splitting workflows

✅ **Expense Logic & Tools**  
- Simplified reconciliation logic  
- Support for fair settlements and net balancing

---

## 🛠️ Getting Started

### Prerequisites
Make sure you have the following installed:

- **Go** (version 1.20+ recommended)
- **Go Modules** (enabled by default in Go 1.16+)
- **MongoDB** (local or cloud instance)

---

### 📥 Installation

```bash
# 1. Clone the repository
git clone https://github.com/kartikey-111/SPLIT-Express-Backend

# 2. Navigate to the project directory
cd SPLIT-Express-Backend

# 3. Install dependencies
go mod tidy
