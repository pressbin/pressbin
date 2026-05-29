package store

import (
	"database/sql"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestAPIKey_HasPermission(t *testing.T) {
	sync := APIKey{Permissions: []string{"posts:write"}}
	if !sync.HasPermission("posts:write") {
		t.Error("expected posts:write")
	}
	if sync.HasPermission("admin") {
		t.Error("sync key should not have admin")
	}

	admin := APIKey{Permissions: []string{"admin"}}
	if !admin.HasPermission("admin") {
		t.Error("expected admin")
	}
	if admin.HasPermission("posts:write") {
		t.Error("admin key should not have posts:write")
	}
}

func TestBootstrap_createsAdminOnce(t *testing.T) {
	st := openTestStore(t)
	raw, err := st.Bootstrap()
	if err != nil {
		t.Fatal(err)
	}
	if raw == "" || len(raw) < len("pb_admin_") {
		t.Fatalf("unexpected bootstrap key: %q", raw)
	}
	raw2, err := st.Bootstrap()
	if err != nil {
		t.Fatal(err)
	}
	if raw2 != "" {
		t.Error("second Bootstrap should return empty string")
	}
	keys, err := st.ListKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatalf("keys count = %d", len(keys))
	}
	if keys[0].Permissions[0] != "admin" {
		t.Errorf("perms = %v", keys[0].Permissions)
	}
}

func TestCreateSyncKey(t *testing.T) {
	st := openTestStore(t)
	raw, err := st.CreateSyncKey("github sync")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(raw, "pb_sync_") {
		t.Fatalf("key = %q", raw)
	}
	found, err := st.FindKeyByRaw(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !found.HasPermission("posts:write") || found.HasPermission("admin") {
		t.Errorf("perms = %v", found.Permissions)
	}
}

func TestFindKeyByRaw(t *testing.T) {
	st := openTestStore(t)
	raw := "pb_sync_" + RandomKeySuffix(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	k := APIKey{
		ID: "testkeyid01", Label: "sync", Prefix: "pb_sync_",
		Hash: string(hash), Permissions: []string{"posts:write"},
	}
	if err := st.CreateKey(k); err != nil {
		t.Fatal(err)
	}
	found, err := st.FindKeyByRaw(raw)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != k.ID {
		t.Errorf("id = %q", found.ID)
	}
	if _, err := st.FindKeyByRaw("pb_sync_wrong"); err == nil {
		t.Error("expected error for wrong key")
	}
	if _, err := st.FindKeyByRaw("not-a-key"); err == nil {
		t.Error("expected error for invalid prefix")
	}
}

func TestRevokeKey(t *testing.T) {
	st := openTestStore(t)
	k := APIKey{
		ID: "revoke01", Label: "x", Prefix: "pb_admin_",
		Hash: "not-used", Permissions: []string{"admin"},
	}
	if err := st.CreateKey(k); err != nil {
		t.Fatal(err)
	}
	if err := st.RevokeKey(k.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.RevokeKey("missing"); err != sql.ErrNoRows {
		t.Errorf("err = %v", err)
	}
}

func TestKeyPrefixForType(t *testing.T) {
	prefix, perms, err := KeyPrefixForType("sync")
	if err != nil || prefix != "pb_sync_" || perms[0] != "posts:write" {
		t.Fatalf("sync: prefix=%q perms=%v err=%v", prefix, perms, err)
	}
	prefix, perms, err = KeyPrefixForType("admin")
	if err != nil || prefix != "pb_admin_" || perms[0] != "admin" {
		t.Fatalf("admin: prefix=%q perms=%v err=%v", prefix, perms, err)
	}
	if _, _, err = KeyPrefixForType("invalid"); err == nil {
		t.Error("expected error")
	}
}

func TestRandomKeySuffix_length(t *testing.T) {
	s := RandomKeySuffix(24)
	if len(s) != 24 {
		t.Errorf("len = %d", len(s))
	}
}
