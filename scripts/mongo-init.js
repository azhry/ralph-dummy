# MongoDB initialization script
db = db.getSiblingDB('wedding_invitations');

// Create application user
db.createUser({
  user: 'wedding_user',
  pwd: 'wedding_password',
  roles: [
    {
      role: 'readWrite',
      db: 'wedding_invitations'
    }
  ]
});

// Create initial collections and indexes
db.createCollection('users');
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "created_at": 1 });

db.createCollection('weddings');
db.weddings.createIndex({ "slug": 1 }, { unique: true });
db.weddings.createIndex({ "user_id": 1 });
db.weddings.createIndex({ "status": 1 });
db.weddings.createIndex({ "is_public": 1 });
db.weddings.createIndex({ "event.date": 1 });
db.weddings.createIndex({ "created_at": 1 });

db.createCollection('rsvps');
db.rsvps.createIndex({ "wedding_id": 1 });
db.rsvps.createIndex({ "email": 1, "wedding_id": 1 });
db.rsvps.createIndex({ "status": 1 });
db.rsvps.createIndex({ "submitted_at": 1 });

db.createCollection('guests');
db.guests.createIndex({ "wedding_id": 1 });
db.guests.createIndex({ "email": 1, "wedding_id": 1 });
db.guests.createIndex({ "rsvp_status": 1 });
db.guests.createIndex({ "side": 1 });
db.guests.createIndex({ "import_batch": 1 });

db.createCollection('analytics');
db.analytics.createIndex({ "wedding_id": 1 });
db.analytics.createIndex({ "event_type": 1 });
db.analytics.createIndex({ "timestamp": 1 });

print('MongoDB initialization completed successfully!');