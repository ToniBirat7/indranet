FROM node:20-alpine AS base
RUN corepack enable && corepack prepare pnpm@9 --activate

# Install deps with lockfile (cached layer)
FROM base AS deps
WORKDIR /app
COPY pnpm-lock.yaml package.json turbo.json ./
COPY packages/shared/package.json ./packages/shared/
COPY packages/web/package.json ./packages/web/
RUN pnpm install --frozen-lockfile --filter @indranet/web...

# Build
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/packages/shared/node_modules ./packages/shared/node_modules
COPY --from=deps /app/packages/web/node_modules ./packages/web/node_modules
COPY packages/shared ./packages/shared
COPY packages/web ./packages/web
COPY package.json turbo.json ./
RUN pnpm --filter @indranet/web build

# Production runner (standalone output)
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/packages/web/.next/standalone ./
COPY --from=builder /app/packages/web/.next/static ./packages/web/.next/static
COPY --from=builder /app/packages/web/public ./packages/web/public
EXPOSE 3000
ENV PORT=3000
CMD ["node", "server.js"]
