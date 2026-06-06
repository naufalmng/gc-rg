package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

const (
	connectSettleDelay = 30 * time.Second
	sendTimeout        = 2 * time.Minute
)

type SendRequest struct {
	JID       string
	PDFPath   string
	Caption   string
	SessionDB string
	FreshAuth bool
	LoginOnly bool
}

func SendReport(ctx context.Context, request SendRequest) error {
	if request.FreshAuth {
		if err := removeSession(request.SessionDB); err != nil {
			return err
		}
	}
	client, err := connectClient(ctx, request.SessionDB)
	if err != nil {
		return err
	}
	defer client.Disconnect()

	fmt.Printf("post_open_settle_seconds=%.0f\n", connectSettleDelay.Seconds())
	select {
	case <-time.After(connectSettleDelay):
	case <-ctx.Done():
		return fmt.Errorf("wait for connection settle: %w", ctx.Err())
	}
	if request.LoginOnly {
		fmt.Println("login_result=connected")
		return nil
	}
	return sendDocument(ctx, client, request)
}

func connectClient(ctx context.Context, sessionDB string) (*whatsmeow.Client, error) {
	if err := os.MkdirAll(filepath.Dir(sessionDB), 0o700); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(ctx, "sqlite3", sqliteDSN(sessionDB), dbLog)
	if err != nil {
		return nil, fmt.Errorf("open session store: %w", err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID != nil {
		fmt.Println("auth_state=existing")
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("connect existing session: %w", err)
		}
		return client, nil
	}

	fmt.Println("auth_state=new")
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("get QR channel: %w", err)
	}
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("connect new session: %w", err)
	}
	for event := range qrChan {
		fmt.Printf("login_event=%s\n", event.Event)
		if event.Event == "code" {
			qrterminal.GenerateHalfBlock(event.Code, qrterminal.L, os.Stdout)
			fmt.Println("auth_qr=scan_with_whatsapp_linked_devices")
		}
		if event.Event == "success" {
			fmt.Println("login_result=paired")
		}
	}
	return client, nil
}

func sendDocument(ctx context.Context, client *whatsmeow.Client, request SendRequest) error {
	targetJID, err := types.ParseJID(request.JID)
	if err != nil {
		return fmt.Errorf("parse WhatsApp JID: %w", err)
	}
	pdfBytes, err := os.ReadFile(request.PDFPath)
	if err != nil {
		return fmt.Errorf("read PDF: %w", err)
	}
	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	uploaded, err := client.Upload(sendCtx, pdfBytes, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("upload PDF: %w", err)
	}
	fileName := filepath.Base(request.PDFPath)
	message := &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
		Title:         proto.String(fileName),
		FileName:      proto.String(fileName),
		Caption:       proto.String(request.Caption),
		Mimetype:      proto.String("application/pdf"),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
	}}
	response, err := client.SendMessage(sendCtx, targetJID, message)
	if err != nil {
		return fmt.Errorf("send WhatsApp document: %w", err)
	}
	fmt.Printf("send_result=sent\n")
	fmt.Printf("message_id=%s\n", response.ID)
	return nil
}

func removeSession(sessionDB string) error {
	if err := os.Remove(sessionDB); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove session DB: %w", err)
	}
	walPath := sessionDB + "-wal"
	shmPath := sessionDB + "-shm"
	_ = os.Remove(walPath)
	_ = os.Remove(shmPath)
	fmt.Printf("fresh_auth=removed %s\n", sessionDB)
	return nil
}

func sqliteDSN(path string) string {
	return "file:" + filepath.ToSlash(path) + "?_foreign_keys=on"
}
