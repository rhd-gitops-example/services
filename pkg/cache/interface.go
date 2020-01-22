package cache

import "context"

type GitCache interface {
	ReadFileFromBranch(ctx context.Context, repoURL, filePath, branch string) ([]byte, error)
	CreateAndCheckoutBranch(ctx context.Context, repoURL, fromBranch, newBranch string) error
	WriteFileToBranchAndStage(ctx context.Context, repoURL, branch, filePath string, data []byte) error
	CommitAndPushBranch(ctx context.Context, repoURL, branch, message, token string) error
}
