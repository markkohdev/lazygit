package git_commands

// "*|feat/detect-purge|origin/feat/detect-purge|[ahead 1]"
import (
	"strconv"
	"testing"
	"time"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/stretchr/testify/assert"
)

func TestObtainBranch(t *testing.T) {
	type scenario struct {
		testName                 string
		input                    []string
		storeCommitDateAsRecency bool
		includeUpstreamStatus    bool
		expectedBranch           *models.Branch
	}

	// Use a time stamp of 2 1/2 hours ago, resulting in a recency string of "2h"
	now := time.Now().Unix()
	timeStamp := strconv.Itoa(int(now - 2.5*60*60))

	scenarios := []scenario{
		{
			testName:                 "TrimHeads",
			input:                    []string{"", "heads/a_branch", "", "", "", "subject", "123", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    true,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "NoUpstream",
			input:                    []string{"", "a_branch", "", "", "", "subject", "123", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    true,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "IsHead",
			input:                    []string{"*", "a_branch", "", "", "", "subject", "123", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    true,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          true,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "IsBehindAndAhead",
			input:                    []string{"", "a_branch", "a_remote/a_branch", "[behind 2, ahead 3]", "[behind 2, ahead 3]", "subject", "123", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    true,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				AheadForPull:  "3",
				BehindForPull: "2",
				AheadForPush:  "3",
				BehindForPush: "2",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "RemoteBranchIsGone",
			input:                    []string{"", "a_branch", "a_remote/a_branch", "[gone]", "[gone]", "subject", "123", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    true,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				UpstreamGone:  true,
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "WithCommitDateAsRecency",
			input:                    []string{"", "a_branch", "", "", "", "subject", "123", timeStamp},
			storeCommitDateAsRecency: true,
			includeUpstreamStatus:    true,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				Recency:       "2h",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "NoTrackBasic",
			input:                    []string{"", "a_branch", "a_remote/a_branch", "subject", "123", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    false,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "123",
			},
		},
		{
			testName:                 "NoTrackNoUpstream",
			input:                    []string{"*", "a_branch", "", "subject", "abc", timeStamp},
			storeCommitDateAsRecency: false,
			includeUpstreamStatus:    false,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          true,
				Subject:       "subject",
				CommitHash:    "abc",
			},
		},
		{
			testName:                 "NoTrackWithRecency",
			input:                    []string{"", "a_branch", "", "subject", "def", timeStamp},
			storeCommitDateAsRecency: true,
			includeUpstreamStatus:    false,
			expectedBranch: &models.Branch{
				Name:          "a_branch",
				Recency:       "2h",
				AheadForPull:  "?",
				BehindForPull: "?",
				AheadForPush:  "?",
				BehindForPush: "?",
				Head:          false,
				Subject:       "subject",
				CommitHash:    "def",
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.testName, func(t *testing.T) {
			branch := obtainBranch(s.input, s.storeCommitDateAsRecency, s.includeUpstreamStatus)
			assert.EqualValues(t, s.expectedBranch, branch)
		})
	}
}

func TestParseUpstreamTrackPatches(t *testing.T) {
	type scenario struct {
		testName string
		input    string
		expected map[string]UpstreamTrackPatch
	}

	scenarios := []scenario{
		{
			testName: "Empty",
			input:    "",
			expected: map[string]UpstreamTrackPatch{},
		},
		{
			testName: "AheadAndBehind",
			input:    "main\x00origin/main\x00[ahead 3, behind 2]\x00[ahead 1]",
			expected: map[string]UpstreamTrackPatch{
				"main": {AheadForPull: "3", BehindForPull: "2", AheadForPush: "1", BehindForPush: "0", UpstreamGone: false},
			},
		},
		{
			testName: "Gone",
			input:    "feature\x00origin/feature\x00[gone]\x00[gone]",
			expected: map[string]UpstreamTrackPatch{
				"feature": {AheadForPull: "?", BehindForPull: "?", AheadForPush: "?", BehindForPush: "?", UpstreamGone: true},
			},
		},
		{
			testName: "NoUpstream",
			input:    "local\x00\x00\x00",
			expected: map[string]UpstreamTrackPatch{
				"local": {AheadForPull: "?", BehindForPull: "?", AheadForPush: "?", BehindForPush: "?", UpstreamGone: false},
			},
		},
		{
			testName: "MultipleBranches",
			input:    "main\x00origin/main\x00\x00\ndev\x00origin/dev\x00[behind 5]\x00",
			expected: map[string]UpstreamTrackPatch{
				"main": {AheadForPull: "0", BehindForPull: "0", AheadForPush: "0", BehindForPush: "0", UpstreamGone: false},
				"dev":  {AheadForPull: "0", BehindForPull: "5", AheadForPush: "0", BehindForPush: "0", UpstreamGone: false},
			},
		},
		{
			testName: "MalformedLineSkipped",
			input:    "only-two-fields\x00origin/foo\nmain\x00origin/main\x00\x00",
			expected: map[string]UpstreamTrackPatch{
				"main": {AheadForPull: "0", BehindForPull: "0", AheadForPush: "0", BehindForPush: "0", UpstreamGone: false},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.testName, func(t *testing.T) {
			result := ParseUpstreamTrackPatches(s.input)
			assert.EqualValues(t, s.expected, result)
		})
	}
}
