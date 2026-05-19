package providers

import "testing"

func TestStripThinkArtifacts_RemovesTaggedReasoning(t *testing.T) {
	input := "<think>internal thoughts</think>public answer"
	got := StripThinkArtifacts(input)
	if got != "public answer" {
		t.Fatalf("expected stripped content, got %q", got)
	}
}

func TestStripThinkArtifacts_CollapsesDuplicateAroundOrphanClose(t *testing.T) {
	input := "who? give me more\n\n</think>\n\nwho? give me more"
	got := StripThinkArtifacts(input)
	if got != "who? give me more" {
		t.Fatalf("expected duplicate collapse, got %q", got)
	}
}

func TestBuildUserVisibleContent_ShowReasoningFalse(t *testing.T) {
	input := "<think>should hide</think>answer"
	got := BuildUserVisibleContent(input, "", false)
	if got != "answer" {
		t.Fatalf("expected answer only, got %q", got)
	}
}

func TestBuildUserVisibleContent_ShowReasoningTrue(t *testing.T) {
	got := BuildUserVisibleContent("answer", "reason details", true)
	expected := "Reasoning:\nreason details\n\nAnswer:\nanswer"
	if got != expected {
		t.Fatalf("expected reasoning+answer block, got %q", got)
	}
}

func TestBuildUserVisibleContent_ExtractsReasoningFromThinkBlocks(t *testing.T) {
	input := "<think>chain</think>final"
	got := BuildUserVisibleContent(input, "", true)
	expected := "Reasoning:\nchain\n\nAnswer:\nfinal"
	if got != expected {
		t.Fatalf("expected extracted reasoning output, got %q", got)
	}
}
