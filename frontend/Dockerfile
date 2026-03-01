FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN npm install
COPY . .
RUN npm run build

FROM joseluisq/static-web-server:2-alpine
COPY --from=builder /app/dist /public
ENV SERVER_PORT=80
ENV SERVER_FALLBACK_PAGE=/public/index.html
ENV SERVER_LOG_LEVEL=info
EXPOSE 80
