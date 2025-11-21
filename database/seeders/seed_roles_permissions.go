package seeders

import (
	"database/sql"
	"fmt"
)

func SeedRolesPermissions(db *sql.DB) error {
	query := `
-- Seed roles
INSERT INTO roles (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Admin', 'Administrator dengan akses penuh')
ON CONFLICT (id) DO NOTHING;

INSERT INTO roles (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440002', 'Mahasiswa', 'Mahasiswa yang melaporkan prestasi')
ON CONFLICT (id) DO NOTHING;

INSERT INTO roles (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440003', 'Dosen Wali', 'Dosen wali yang memverifikasi prestasi')
ON CONFLICT (id) DO NOTHING;

-- Seed permissions
INSERT INTO permissions (id, name, resource, action, description) VALUES
('650e8400-e29b-41d4-a716-446655440001', 'achievement:create', 'achievement', 'create', 'Membuat prestasi baru')
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (id, name, resource, action, description) VALUES
('650e8400-e29b-41d4-a716-446655440002', 'achievement:read', 'achievement', 'read', 'Melihat prestasi')
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (id, name, resource, action, description) VALUES
('650e8400-e29b-41d4-a716-446655440003', 'achievement:update', 'achievement', 'update', 'Mengubah prestasi')
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (id, name, resource, action, description) VALUES
('650e8400-e29b-41d4-a716-446655440004', 'achievement:delete', 'achievement', 'delete', 'Menghapus prestasi')
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (id, name, resource, action, description) VALUES
('650e8400-e29b-41d4-a716-446655440005', 'achievement:verify', 'achievement', 'verify', 'Memverifikasi prestasi')
ON CONFLICT (id) DO NOTHING;

INSERT INTO permissions (id, name, resource, action, description) VALUES
('650e8400-e29b-41d4-a716-446655440006', 'user:manage', 'user', 'manage', 'Mengelola user')
ON CONFLICT (id) DO NOTHING;

-- Admin: full access
INSERT INTO role_permissions (role_id, permission_id) VALUES
('550e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001'),
('550e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440002'),
('550e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440003'),
('550e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440004'),
('550e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440005'),
('550e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440006')
ON CONFLICT DO NOTHING;

-- Mahasiswa
INSERT INTO role_permissions (role_id, permission_id) VALUES
('550e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440001'),
('550e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440002'),
('550e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440003')
ON CONFLICT DO NOTHING;

-- Dosen Wali
INSERT INTO role_permissions (role_id, permission_id) VALUES
('550e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440002'),
('550e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440005')
ON CONFLICT DO NOTHING;
`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("seeding failed: %v", err)
	}

	fmt.Println("Seeder seed_roles_permissions executed successfully")
	return nil
}
