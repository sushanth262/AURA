package rag_test

import (
	"context"
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/rag"
)

func TestStubClient_TenantIsolation(t *testing.T) {
	client := rag.StubClient{}
	demo, err := client.Retrieve(context.Background(), orchestration.RAGQuery{
		TenantID:   "demo",
		Namespaces: []string{"runbooks"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, hit := range demo {
		if hit["tenant_id"] != "demo" {
			t.Fatalf("leaked tenant: %v", hit)
		}
		if hit["namespace"] != "runbooks" {
			t.Fatalf("wrong namespace: %v", hit)
		}
	}

	other, err := client.Retrieve(context.Background(), orchestration.RAGQuery{
		TenantID:   "other",
		Namespaces: []string{"runbooks"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, hit := range other {
		if hit["tenant_id"] == "demo" {
			t.Fatal("demo tenant doc leaked to other tenant")
		}
	}
}

func TestStubClient_NamespaceFilter(t *testing.T) {
	client := rag.StubClient{}
	hits, err := client.Retrieve(context.Background(), orchestration.RAGQuery{
		TenantID:   "demo",
		Namespaces: []string{"source_code"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected source_code hit")
	}
	for _, h := range hits {
		if h["namespace"] != "source_code" {
			t.Fatalf("got namespace %v", h["namespace"])
		}
	}
}
