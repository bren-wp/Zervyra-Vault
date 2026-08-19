package core

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"
)

const (
	Format                   = "BRENDIGO_VAULT_NATIVE_V1"
	KDFIterations            = 600000
	MinKDFIterations         = 100000
	MaxKDFIterations         = 2000000
	MaxVaultFileSize         = 64 << 20
	MaxEntries               = 100000
	BackupCount              = 10
	ImmediateSnapshotCount   = 3
	BackupCheckpointInterval = 5 * time.Minute
	LockStaleAfter           = 2 * time.Minute
	DefaultPasswordLen       = 24
	MaxEntryRevisions        = 20
)

type PasswordHistoryItem struct { Password string `json:"password"`; ChangedAt string `json:"changedAt"` }
type EntryRevision struct { Title string `json:"title"`; Username string `json:"username"`; Email string `json:"email,omitempty"`; Password string `json:"password"`; URL string `json:"url"`; Notes string `json:"notes"`; TOTP string `json:"totp"`; Tags string `json:"tags,omitempty"`; Favorite bool `json:"favorite"`; ChangedAt string `json:"changedAt"` }
type Entry struct { ID string `json:"id"`; Title string `json:"title"`; Username string `json:"username"`; Email string `json:"email,omitempty"`; Password string `json:"password"`; URL string `json:"url"`; Notes string `json:"notes"`; TOTP string `json:"totp"`; Tags string `json:"tags,omitempty"`; Favorite bool `json:"favorite"`; CreatedAt string `json:"createdAt,omitempty"`; UpdatedAt string `json:"updatedAt,omitempty"`; PasswordHistory []PasswordHistoryItem `json:"passwordHistory,omitempty"`; Revisions []EntryRevision `json:"revisions,omitempty"` }
type Vault struct { Name string `json:"name"`; Entries []Entry `json:"entries"`; Trash []Entry `json:"trash,omitempty"`; UpdatedAt string `json:"updatedAt"` }
type envelope struct { Format string `json:"format"`; Iterations int `json:"iterations"`; Salt string `json:"salt"`; Nonce string `json:"nonce"`; Payload string `json:"payload"` }
type VaultLock struct{ path string }

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339Nano) }
func RandomID() string { b:=make([]byte,16); if _,err:=io.ReadFull(rand.Reader,b); err!=nil { return fmt.Sprintf("fallback-%d",time.Now().UnixNano()) }; return fmt.Sprintf("%x",b) }
func NewVault() Vault { return Vault{Name:"Zervyra Vault", Entries:[]Entry{}, Trash:[]Entry{}, UpdatedAt:nowRFC3339()} }
func NewEntry() Entry { now:=nowRFC3339(); return Entry{ID:RandomID(),CreatedAt:now,UpdatedAt:now,PasswordHistory:[]PasswordHistoryItem{},Revisions:[]EntryRevision{}} }
func SnapshotEntry(e Entry) EntryRevision { return EntryRevision{Title:e.Title,Username:e.Username,Email:e.Email,Password:e.Password,URL:e.URL,Notes:e.Notes,TOTP:e.TOTP,Tags:e.Tags,Favorite:e.Favorite,ChangedAt:nowRFC3339()} }
func RevisionMatchesEntry(r EntryRevision,e Entry) bool { return r.Title==e.Title&&r.Username==e.Username&&r.Email==e.Email&&r.Password==e.Password&&r.URL==e.URL&&r.Notes==e.Notes&&r.TOTP==e.TOTP&&r.Tags==e.Tags&&r.Favorite==e.Favorite }
func AppendRevision(e *Entry,r EntryRevision){ if e==nil||RevisionMatchesEntry(r,*e){return}; r.ChangedAt=nowRFC3339(); e.Revisions=append(e.Revisions,r); if len(e.Revisions)>MaxEntryRevisions { e.Revisions=append([]EntryRevision(nil),e.Revisions[len(e.Revisions)-MaxEntryRevisions:]...) } }
func RestoreLastRevision(e *Entry) bool { if e==nil||len(e.Revisions)==0{return false}; last:=e.Revisions[len(e.Revisions)-1]; e.Revisions=e.Revisions[:len(e.Revisions)-1]; e.Title,e.Username,e.Email,e.Password=last.Title,last.Username,last.Email,last.Password; e.URL,e.Notes,e.TOTP,e.Tags,e.Favorite=last.URL,last.Notes,last.TOTP,last.Tags,last.Favorite; e.UpdatedAt=nowRFC3339(); return true }
func CloneVault(v Vault) Vault { out:=v; cloneEntries:=func(src []Entry)[]Entry{ dst:=make([]Entry,len(src)); copy(dst,src); for i:=range dst { if src[i].PasswordHistory!=nil {dst[i].PasswordHistory=append([]PasswordHistoryItem(nil),src[i].PasswordHistory...)}; if src[i].Revisions!=nil {dst[i].Revisions=append([]EntryRevision(nil),src[i].Revisions...)} }; return dst }; out.Entries=cloneEntries(v.Entries); out.Trash=cloneEntries(v.Trash); return out }
