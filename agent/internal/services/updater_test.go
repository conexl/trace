package services

import (
	"context"
	"errors"
	"testing"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/logger"
	"agent/internal/updater"
)

type fakeUpdater struct {
	applyResult updater.Result
	applyErr    error
	checkResult updater.Result
	checkErr    error
	applyCalls  int
	checkCalls  int
}

func (f *fakeUpdater) ApplyOptions(ctx context.Context, opts updater.Options, target string) (updater.Result, error) {
	f.applyCalls++
	return f.applyResult, f.applyErr
}

func (f *fakeUpdater) CheckOptions(ctx context.Context, opts updater.Options) (updater.Result, error) {
	f.checkCalls++
	return f.checkResult, f.checkErr
}

type fakeBufferedSink struct {
	published []collectors.Snapshot
}

func (f *fakeBufferedSink) PublishSnapshot(ctx context.Context, snapshot collectors.Snapshot) error {
	f.published = append(f.published, snapshot)
	return nil
}

func (f *fakeBufferedSink) ReadBatch(limit int) ([]collectors.Snapshot, error) { return nil, nil }
func (f *fakeBufferedSink) Ack(count int) error                                 { return nil }
func (f *fakeBufferedSink) Close() error                                        { return nil }
func (f *fakeBufferedSink) Count() int                                          { return len(f.published) }

func TestCheckUpdateManualDoesNothing(t *testing.T) {
	u := &fakeUpdater{}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "manual", URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done || u.applyCalls != 0 || u.checkCalls != 0 {
		t.Fatalf("unexpected result: done=%v err=%v apply=%d check=%d", done, err, u.applyCalls, u.checkCalls)
	}
}

func TestCheckUpdateEmptyURLDoesNothing(t *testing.T) {
	u := &fakeUpdater{}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "check"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done || u.applyCalls != 0 || u.checkCalls != 0 {
		t.Fatalf("unexpected result: done=%v err=%v apply=%d check=%d", done, err, u.applyCalls, u.checkCalls)
	}
}

func TestCheckUpdateNilUpdaterDoesNothing(t *testing.T) {
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "check", URL: "http://example.com/update"}},
		updater: nil,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done {
		t.Fatalf("unexpected result: done=%v err=%v", done, err)
	}
}

func TestCheckUpdateDefaultPolicyIsCheck(t *testing.T) {
	u := &fakeUpdater{checkResult: updater.Result{SHA256: "abc123"}}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done || u.checkCalls != 1 {
		t.Fatalf("unexpected result: done=%v err=%v check=%d", done, err, u.checkCalls)
	}
}

func TestCheckUpdateCheckPolicyReportsAvailable(t *testing.T) {
	buf := &fakeBufferedSink{}
	u := &fakeUpdater{checkResult: updater.Result{SHA256: "newsha", SignatureVerified: true}}
	a := &Agent{
		cfg:     config.Config{Agent: config.AgentConfig{Name: "test"}, Update: config.UpdateConfig{Policy: "check", URL: "http://example.com/update"}},
		updater: u,
		buffer:  buf,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done {
		t.Fatalf("unexpected result: done=%v err=%v", done, err)
	}
	if len(buf.published) != 1 || len(buf.published[0].Events) != 1 {
		t.Fatalf("expected one published event, got %#v", buf.published)
	}
	ev := buf.published[0].Events[0]
	if ev.Type != "update.available" {
		t.Fatalf("unexpected event type %q", ev.Type)
	}
}

func TestCheckUpdateCheckPolicySameSHAReportsNothing(t *testing.T) {
	current, err := updater.CurrentExecutableSHA256()
	if err != nil {
		t.Fatalf("current executable hash: %v", err)
	}
	buf := &fakeBufferedSink{}
	u := &fakeUpdater{checkResult: updater.Result{SHA256: current}}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "check", URL: "http://example.com/update"}},
		updater: u,
		buffer:  buf,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done {
		t.Fatalf("unexpected result: done=%v err=%v", done, err)
	}
	if len(buf.published) != 0 {
		t.Fatalf("expected no events, got %#v", buf.published)
	}
}

func TestCheckUpdateCheckPolicyReturnsError(t *testing.T) {
	u := &fakeUpdater{checkErr: errors.New("download failed")}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "check", URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err == nil || done {
		t.Fatalf("expected error, got done=%v err=%v", done, err)
	}
}

func TestCheckUpdateAutoPolicyAppliesAndExits(t *testing.T) {
	exited := 0
	oldExit := osExit
	osExit = func(code int) { exited = code }
	defer func() { osExit = oldExit }()

	u := &fakeUpdater{applyResult: updater.Result{Updated: true, SHA256: "newsha", SignatureVerified: true}}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "auto", URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done || u.applyCalls != 1 || u.checkCalls != 0 {
		t.Fatalf("unexpected result: done=%v err=%v apply=%d check=%d", done, err, u.applyCalls, u.checkCalls)
	}
	if exited != 0 {
		t.Fatalf("expected exit code 0, got %d", exited)
	}
}

func TestCheckUpdateAutoPolicySameSHADoesNotExit(t *testing.T) {
	current, err := updater.CurrentExecutableSHA256()
	if err != nil {
		t.Fatalf("current executable hash: %v", err)
	}
	exited := -1
	oldExit := osExit
	osExit = func(code int) { exited = code }
	defer func() { osExit = oldExit }()

	u := &fakeUpdater{applyResult: updater.Result{Updated: true, SHA256: current}}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "auto", URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done || u.applyCalls != 1 {
		t.Fatalf("unexpected result: done=%v err=%v apply=%d", done, err, u.applyCalls)
	}
	if exited != -1 {
		t.Fatalf("unexpected exit code %d", exited)
	}
}

func TestCheckUpdateAutoPolicyReturnsApplyError(t *testing.T) {
	u := &fakeUpdater{applyErr: errors.New("apply failed")}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "auto", URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err == nil || done {
		t.Fatalf("expected error, got done=%v err=%v", done, err)
	}
}

func TestCheckUpdateAutoPolicyNotUpdatedDoesNothing(t *testing.T) {
	u := &fakeUpdater{applyResult: updater.Result{Updated: false}}
	a := &Agent{
		cfg:     config.Config{Update: config.UpdateConfig{Policy: "auto", URL: "http://example.com/update"}},
		updater: u,
	}
	done, err := a.checkUpdate(context.Background())
	if err != nil || done || u.applyCalls != 1 {
		t.Fatalf("unexpected result: done=%v err=%v apply=%d", done, err, u.applyCalls)
	}
}

var _ logger.BufferedSink = (*fakeBufferedSink)(nil)
