FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM node:20-alpine

WORKDIR /app

COPY --from=builder /app/dist /app/dist
COPY package.json package-lock.json ./
RUN npm ci --omit=dev

EXPOSE 4000

ENV PORT=4000
ENV NODE_ENV=production

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
    CMD node -e "require('http').get('http://localhost:4000/health', (res) => { if (res.statusCode === 200) process.exit(0); else process.exit(1); }).on('error', () => process.exit(1))"

CMD ["node", "dist/ui/server/server.mjs"]
