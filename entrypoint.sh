#!/bin/sh
set -e

echo "Running migrations..."
./bin/migrate

echo "Seeding FAQs..."
./bin/seed-faq

echo "Starting moderation worker..."
./bin/moderation &

echo "Starting API server..."
exec ./bin/api
