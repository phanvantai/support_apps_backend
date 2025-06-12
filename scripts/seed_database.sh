#!/bin/bash

# Database seeding script for the Support App Backend

set -e

echo "🌱 Seeding Support App Backend Database"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration - Update these for your environment
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${DB_NAME:-support_app}

# Check if PostgreSQL is available
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Function to execute SQL
execute_sql() {
    local sql="$1"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$sql"
}

echo "📡 Connecting to database..."
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null; then
    echo "❌ Failed to connect to database. Please ensure PostgreSQL is running and credentials are correct."
    exit 1
fi

echo -e "${GREEN}✅ Connected to database${NC}"

echo "🗑️  Cleaning existing data..."
execute_sql "DELETE FROM support_requests;"

echo "📝 Inserting sample support requests..."

# Sample support requests
execute_sql "
INSERT INTO support_requests (type, user_email, message, platform, app_version, device_model, status, admin_notes, created_at, updated_at) VALUES
('support', 'john.doe@example.com', 'I cannot login to my account. I keep getting an invalid credentials error even though I am sure my password is correct.', 'iOS', '2.1.0', 'iPhone 14 Pro', 'new', NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
('feedback', 'sarah.wilson@example.com', 'Love the new dark mode feature! The app looks much better now. Could you add more customization options?', 'Android', '2.0.5', 'Samsung Galaxy S23', 'resolved', 'Thank you for the feedback! We are working on more customization options for the next release.', NOW() - INTERVAL '1 day', NOW() - INTERVAL '6 hours'),
('support', 'mike.johnson@example.com', 'The app crashes every time I try to upload a photo. This has been happening since the last update.', 'iOS', '2.1.0', 'iPhone 13', 'in_progress', 'We have identified the issue and are working on a fix. Will be included in version 2.1.1.', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day'),
('feedback', NULL, 'Great app overall, but the loading times could be improved. Sometimes it takes 10+ seconds to load the main screen.', 'Android', '2.0.4', 'Google Pixel 7', 'new', NULL, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours'),
('support', 'alice.brown@example.com', 'I accidentally deleted an important item and cannot find it in the trash. Is there a way to recover it?', 'iOS', '2.0.8', 'iPad Air', 'resolved', 'Items are kept in our backup for 30 days. We have restored your item.', NOW() - INTERVAL '1 week', NOW() - INTERVAL '3 days'),
('support', 'robert.garcia@example.com', 'Push notifications are not working on my device. I have checked all the settings and they seem correct.', 'Android', '2.1.0', 'OnePlus 11', 'new', NULL, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes'),
('feedback', 'emma.davis@example.com', 'The new search feature is amazing! It is so much faster than before. Keep up the good work!', 'iOS', '2.1.0', 'iPhone 12 Mini', 'resolved', 'Thank you for the positive feedback!', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
('support', 'david.lee@example.com', 'Cannot sync data between my phone and tablet. Both devices show different information.', 'Android', '2.0.9', 'Samsung Galaxy Tab S8', 'in_progress', 'Our team is investigating sync issues. We will update you soon.', NOW() - INTERVAL '4 days', NOW() - INTERVAL '2 days'),
('feedback', NULL, 'Would love to see a widget for the home screen. That would make accessing the app much more convenient.', 'iOS', '2.0.7', 'iPhone SE', 'new', NULL, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
('support', 'lisa.martinez@example.com', 'The export feature is not working. When I try to export my data, I get an error message saying \"Export failed\".', 'Android', '2.1.0', 'Xiaomi 13', 'new', NULL, NOW() - INTERVAL '15 minutes', NOW() - INTERVAL '15 minutes');
"

echo "📊 Checking inserted data..."
count=$(execute_sql "SELECT COUNT(*) FROM support_requests;" | grep -o '[0-9]*' | head -1)
echo -e "${GREEN}✅ Successfully inserted $count support requests${NC}"

echo ""
echo "📋 Sample data summary:"
execute_sql "
SELECT 
    type,
    status,
    COUNT(*) as count
FROM support_requests 
GROUP BY type, status 
ORDER BY type, status;
"

echo ""
echo -e "${GREEN}🎉 Database seeding completed successfully!${NC}"
echo ""
echo "You can now test the API with the following sample data:"
echo "- 6 support requests and 4 feedback requests"
echo "- Various statuses: new, in_progress, resolved"
echo "- Mixed platforms: iOS and Android"
echo "- Some with admin notes, some without"
