package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func writeAtomicFile(path string, raw []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil { return err }
	tmp, err := os.CreateTemp(dir, ".zervyra-vault-*.tmp")
	if err != nil { return err }
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil { tmp.Close(); return err }
	if _, err := tmp.Write(raw); err != nil { tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	return atomicReplace(tmpName, path)
}

func Save(path, password string, v Vault) error { return saveInternal(path,password,v,true) }
func SaveAutosave(path,password string,v Vault) error { rotate:=false; if st,err:=os.Stat(filepath.Clean(strings.TrimSpace(path))+".bak1"); err!=nil || time.Since(st.ModTime())>=BackupCheckpointInterval {rotate=true}; return saveInternal(path,password,v,rotate) }

func saveInternal(path,password string,v Vault,rotate bool) error {
	path=filepath.Clean(strings.TrimSpace(path)); if path=="."||path=="" {return errors.New("invalid vault path")}
	raw,err:=encodeVault(password,v); if err!=nil{return err}
	recovery:=path+".recovery"
	if err:=writeAtomicFile(recovery,raw);err!=nil{return fmt.Errorf("recovery write failed: %w",err)}
	if rr,err:=os.ReadFile(recovery);err!=nil{return fmt.Errorf("recovery verification read failed: %w",err)} else if _,err:=decodeVault(rr,password);err!=nil{return fmt.Errorf("recovery verification failed: %w",err)}
	currentValid:=false; if _,err:=os.Stat(path);err==nil{if _,loadErr:=Load(path,password);loadErr==nil{currentValid=true}}
	if currentValid { if err:=rotateImmediateSnapshots(path,ImmediateSnapshotCount);err!=nil{return fmt.Errorf("immediate snapshot failed: %w",err)}; if rotate {if err:=rotateBackups(path,BackupCount);err!=nil{return fmt.Errorf("backup failed: %w",err)}} }
	if err:=writeAtomicFile(path,raw);err!=nil{return fmt.Errorf("save failed (newest data remains in .recovery): %w",err)}
	check,err:=os.ReadFile(path); if err!=nil{return fmt.Errorf("post-save verification read failed: %w",err)}; if _,err:=decodeVault(check,password);err!=nil{_ = copyFile(recovery,path); return fmt.Errorf("post-save verification failed; recovery copy restored: %w",err)}
	return nil
}

func Load(path,password string)(Vault,error){ var v Vault; path=filepath.Clean(strings.TrimSpace(path)); st,err:=os.Stat(path); if err!=nil{return v,err}; if st.Size()<=0||st.Size()>MaxVaultFileSize{return v,errors.New("vault file size is invalid")}; raw,err:=os.ReadFile(path); if err!=nil{return v,err}; return decodeVault(raw,password) }
type RecoveryResult struct { Vault Vault; Source string; Recovered bool }
func LoadBest(path,password string)(RecoveryResult,error){ path=filepath.Clean(strings.TrimSpace(path)); type generation struct{path string;v Vault;when time.Time}; candidates:=[]string{path,path+".recovery"}; for i:=1;i<=ImmediateSnapshotCount;i++{candidates=append(candidates,fmt.Sprintf("%s.prev%d",path,i))}; for i:=1;i<=BackupCount;i++{candidates=append(candidates,fmt.Sprintf("%s.bak%d",path,i))}; valid:=make([]generation,0,len(candidates)); var mainErr error; for _,candidate:=range candidates{v,err:=Load(candidate,password);if err!=nil{if candidate==path{mainErr=err};continue};when:=time.Time{};if parsed,err:=time.Parse(time.RFC3339Nano,v.UpdatedAt);err==nil{when=parsed}else if st,err:=os.Stat(candidate);err==nil{when=st.ModTime()};valid=append(valid,generation{candidate,v,when})};if len(valid)==0{if mainErr!=nil{return RecoveryResult{},mainErr};return RecoveryResult{},errors.New("no valid vault generation could be recovered")};best:=valid[0];for _,g:=range valid[1:]{if g.when.After(best.when){best=g}};return RecoveryResult{Vault:best.v,Source:best.path,Recovered:best.path!=path},nil }
func rotateImmediateSnapshots(path string,count int)error{if count<=0{return nil};for i:=count;i>=2;i--{src:=fmt.Sprintf("%s.prev%d",path,i-1);dst:=fmt.Sprintf("%s.prev%d",path,i);if _,err:=os.Stat(src);err==nil{if err:=copyFile(src,dst);err!=nil{return err}}};return copyFile(path,path+".prev1")}
func rotateBackups(path string,count int)error{if count<=0{return nil};for i:=count;i>=2;i--{src:=fmt.Sprintf("%s.bak%d",path,i-1);dst:=fmt.Sprintf("%s.bak%d",path,i);if _,err:=os.Stat(src);err==nil{if err:=copyFile(src,dst);err!=nil{return err}}};return copyFile(path,path+".bak1")}
func copyFile(src,dst string)error{in,err:=os.Open(src);if err!=nil{return err};defer in.Close();tmp:=dst+".tmp";out,err:=os.OpenFile(tmp,os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0600);if err!=nil{return err};ok:=false;defer func(){out.Close();if !ok{os.Remove(tmp)}}();if _,err:=io.Copy(out,in);err!=nil{return err};if err:=out.Sync();err!=nil{return err};if err:=out.Close();err!=nil{return err};if err:=atomicReplace(tmp,dst);err!=nil{return err};ok=true;return nil}
func readLockPID(lockPath string)int{b,err:=os.ReadFile(lockPath);if err!=nil{return 0};for _,line:=range strings.Split(string(b),"\n"){if !strings.HasPrefix(line,"pid="){continue};pid,err:=strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line,"pid=")));if err==nil&&pid>0{return pid}};return 0}
