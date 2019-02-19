package main

import (
	"fmt"
	"math/rand"
	"time"

	cedar "github.com/iohub/Ahocorasick"
)

// BlockedUser 被拉黑用户数据结构
type BlockedUser struct {
	userID   string
	expireAt int64 // Unix时间戳，0表示用不过期
}

// initBlockedUsers 初始化黑名单用户列表
func initBlockedUsers() *[]BlockedUser {
	return &[]BlockedUser{
		{"16117700", time.Now().Unix() + 3600},
		{"11597870", time.Now().Unix() + 3600},
		{"27555600", time.Now().Unix() + 3600},
		{"23627518", time.Now().Unix() + 3600},
		{"28180817", time.Now().Unix() + 3600},
		{"13261790", time.Now().Unix() + 3600},
		{"32003775", time.Now().Unix() + 3600},
		{"22704653", time.Now().Unix() + 3600},
		{"18551900", time.Now().Unix() + 3600},
		{"23856909", time.Now().Unix() + 3600},
		{"21340345", time.Now().Unix() + 3600},
		{"22892796", time.Now().Unix() + 3600},
		{"11987660", time.Now().Unix() + 3600},
		{"23901468", time.Now().Unix() + 3600},
		{"23462516", time.Now().Unix() + 3600},
		{"36506849", time.Now().Unix() + 3600},
	}
}

// isBlocked 判断用户是否被拉黑
func isBlocked(userID string, trie *cedar.Cedar) bool {
	val, err := trie.Get([]byte(userID))
	if err != nil {
		// 用户未被拉黑
		return false
	}
	return time.Now().Unix() <= val.(int64)
}

// showUserStatus 显示用户当前状态
func showUserStatus(userID string, trie *cedar.Cedar) {
	if blocked := isBlocked(userID, trie); blocked {
		fmt.Println("User", userID, "has been blocked")
	} else {
		fmt.Println("User", userID, "not in blocklist.")
	}
}

// newTrie 根据初始化拉黑用户列表构建 Trie
func newTrie(blockedUsers *[]BlockedUser) *cedar.Cedar {
	trie := cedar.NewCedar()
	for _, bu := range *blockedUsers {
		trie.Insert([]byte(bu.userID), bu.expireAt)
	}
	return trie
}

// addBlockUser 将用户添加至黑名单
func addBlockUser(userID string, trie *cedar.Cedar) {
	fmt.Println("Adding user", userID, "to blacklist...")
	trie.Insert([]byte(userID), time.Now().Unix()+int64(rand.Intn(10000)))
}

// delBlockUser 从黑名单中删除用户
func delBlockUser(userID string, trie *cedar.Cedar) {
	fmt.Println("Deleting user", userID, "from blacklist...")
	trie.Delete([]byte(userID))
}

func main() {
	blockedUsers := initBlockedUsers()
	trie := newTrie(blockedUsers)
	fmt.Printf("The trie has been created with %d blocked users.\n", len(*blockedUsers))

	// 检查下列用户是否被黑名单屏蔽
	testUsers := []string{
		"16117700", "499123", "161177",
	}

	for _, user := range testUsers {
		showUserStatus(user, trie)
	}

	// 逐个添加新的用户至黑名单
	addBlockUser("susan", trie)
	addBlockUser("jimmy", trie)
	addBlockUser("frank", trie)
	addBlockUser("monic", trie)

	showUserStatus("jimmy", trie)
	showUserStatus("x-man", trie)

	// 从黑名单中删除 jimmy 用户
	delBlockUser("jimmy", trie)
	showUserStatus("jimmy", trie)
}
