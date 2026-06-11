FROM node:20-alpine AS base
RUN corepack enable && corepack prepare pnpm@9 --activate

# Install deps with lockfile (cached layer)
FROM base AS deps
WORKDIR /app
COPY pnpm-lock.yaml package.json pnpm-workspace.yaml turbo.json ./
COPY packages/shared/package.json ./packages/shared/
COPY packages/web/package.json ./packages/web/
RUN pnpm install --frozen-lockfile --filter @indranet/web...

# Build
FROM base AS builder
WORKDIR /app
COPY pnpm-workspace.yaml package.json turbo.json ./
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/packages/shared/node_modules ./packages/shared/node_modules
COPY --from=deps /app/packages/web/node_modules ./packages/web/node_modules
COPY packages/shared ./packages/shared
COPY packages/web ./packages/web
RUN pnpm --filter @indranet/web build

# Production runner (standalone output)
# standalone/ contains server.js at its root; the server uses __dirname to locate
# static assets, so they must be at .next/static and public/ relative to server.js.
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/packages/web/.next/standalone ./
COPY --from=builder /app/packages/web/.next/static ./.next/static
COPY --from=builder /app/packages/web/public ./public
EXPOSE 3000
ENV PORT=3000
CMD ["node", "server.js"]
