# Go Web File Share

This is a simple file-sharing web application built in Go. It acts as a central relay server where a user can upload a file, receive a unique 6-character code, and another user (on the same network) can download that file using the code.

This project demonstrates a shift from a command-line P2P app to a browser-based "broker" model, which is necessary because web browsers cannot act as servers for security reasons.

## ✨ Features

  * **Clean Web UI:** A simple, responsive, and modern "blue spectrum" UI with "Send" and "Receive" tabs.
  * **File Upload:** Users can upload any file. The server handles it as multipart form data.
  * **Unique Share Code:** A 6-character, random, and collision-resistant share code is generated for each upload.
  * **File Download:** Users can download the file simply by entering the code.
  * **Concurrency-Safe:** The server uses a `sync.RWMutex` to safely handle the file map, allowing many users to upload and download at the same time.

## 🛠️ Tech Stack

  * **Backend:** Go (Golang)
      * `net/http`: For the webserver, file serving, and API endpoints.
      * `sync`: For the `RWMutex` to protect the file store.
      * `crypto/rand`: For generating a secure random share code.
  * **Frontend:** Vanilla HTML, CSS, and JavaScript (no frameworks).
  * **Storage:** The server's local file system (in the `/uploads` folder).

## 🚀 How to Run

### Prerequisites

  * Go (Version 1.18 or newer)

### 1\. Initialize Your Go Module

In your project's directory, run this command to create a `go.mod` file:

```bash
go mod init web-share
```

### 2\. Create the Uploads Directory

This is a **mandatory step**. The Go server needs a place to temporarily store the uploaded files.

```bash
mkdir uploads
```

### 3\. Run the Server

Make sure your `main.go`, `index.html`, and `uploads` folder are all in the same directory. Then, run the application:

```bash
go run .
```

You will see the following output, and your service will be running:

```bash
Web share server starting on :8080...
Open http://localhost:8080 to share a file.
```

### 4\. Open the App

Open your web browser and go to:

**`http://localhost:8080/`**

-----

## 📝 How to Use

### To Send a File

1.  Open `http://localhost:8080` in your browser.
2.  On the "Send" tab, click "Choose File" and select the file you want to share.
3.  Click the "Upload and Get Code" button.
4.  Wait for the upload to complete. A 6-character code (e.g., `a4bc12`) will appear.
5.  Send this code to your friend.

### To Receive a File

1.  Open `http://localhost:8080` in a browser.
2.  Click the "Receive" tab.
3.  Enter the 6-character code your friend sent you.
4.  Click "Download File."
5.  Your browser will download the file.

-----

## 🧠 Architecture & Key Concepts

  * **Broker/Relay Model:** A web browser cannot listen for incoming TCP connections (like our P2P app did). Therefore, the Go application must act as a central "broker" or "relay." The sender **uploads** the file *to* the server, and the receiver **downloads** the file *from* the server.
  * **`FileStore` (map + mutex):** The server uses a `map[string]string` to associate the random share code (e.g., `"a4bc12"`) with the file's path on the server (e.g., `"uploads/share-12345-myfile.zip"`). This map is wrapped in a `sync.RWMutex` to prevent race conditions when multiple people upload or download simultaneously.
  * **File Handling:**
      * **Upload:** Uses `r.ParseMultipartForm` and `io.Copy` to stream the uploaded file into a *temporary file* in the `./uploads` directory.
      * **Download:** Uses `http.ServeFile` to efficiently stream the file from the server to the receiver. It sets the `Content-Disposition` header to tell the browser to "download" the file rather than trying to "display" it.

## ⚠️ Disclaimer

This is a demo project and is **not** intended for production use.

**No Cleanup Logic:** Uploaded files are stored in the `/uploads` folder and are **never deleted**. In a real-world application, you would need to add a cleanup mechanism (e.g., a background goroutine) to delete files after a certain time (like 1 hour) to prevent the server's disk from filling up.
