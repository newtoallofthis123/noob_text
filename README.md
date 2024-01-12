# NoobText

Changing the way I think about servers.

So, I don't consider myself to be like a super cool dev who knows exactly what they want and what is best for the project. I am still discovering new things, and I am still learning.

All my web journey, I was always in the server to client paradigm. The server sends data in JSON format the the client interprets with the help of heavy dom manipulating libraries like React, Angular, Vue, etc.

However, HTMX changed the way I think about servers. I am not saying that I am going to use it in all my projects, but I am going to use it for this one and try to implement the following in the project using

- Go (Golang)
- HTMX
- TailwindCSS
- PostgreSQL

Here's all the things I want to implement in this project:
(I might have already implemented a few)

- [x] CRUD operations
- [x] Authentication
- [x] Dynamic routing
- [x] Redis Cache

What all are done? None for now. I am just starting out. I will update this as I go along.

## Installation

```bash
git clone https://github.com/newtoallofthis123/noob_text/
cd noob_text
make run
```

You also need tailwindcss cli to compile the css. You can install it using npm.

```bash
npm install -D tailwindcss
```

Then, you have watch the `public/input.css` file and compile it to `public/output.css` file.

```bash
npx tailwindcss -i public/input.css -o public/output.css --watch
```

You will find the compiled css in `public/output.css` file.

Moreover, you will need to set all the environment variables in the `.env` file. You can use the `.env.example` file as a reference.
