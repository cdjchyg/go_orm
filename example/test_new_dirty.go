package main

import (
	"fmt"

	"DB/example/pb"
)

func main() {
	fmt.Println("=== 测试新的脏标记功能 ===")

	// 创建User实例
	user := pb.NewUser()

	// 测试初始状态
	fmt.Printf("初始状态: IsDirty=%v, DirtyCount=%d\n", user.IsDirty(), user.GetDirtyFieldCount())

	// 设置一些字段
	user.SetName("Alice")
	user.SetAge(25)

	fmt.Printf("设置Name和Age后: IsDirty=%v, DirtyCount=%d\n", user.IsDirty(), user.GetDirtyFieldCount())
	fmt.Printf("脏字段索引: %v\n", user.GetDirtyFieldIndexes())
	fmt.Printf("Name是否脏: %v, Age是否脏: %v\n", user.IsNameDirty(), user.IsAgeDirty())
	fmt.Printf("Email是否脏: %v, Id是否脏: %v\n", user.IsEmailDirty(), user.IsIdDirty())

	// 测试嵌套对象
	profile := pb.NewUserProfile()
	profile.SetAvatarUrl("http://example.com/avatar.jpg")
	user.SetProfile(profile)

	fmt.Printf("设置Profile后: IsDirty=%v, DirtyCount=%d\n", user.IsDirty(), user.GetDirtyFieldCount())
	fmt.Printf("Profile是否脏: %v\n", user.IsProfileDirty())

	// 测试嵌套对象的脏标记传播
	userProfile := user.GetProfile()
	userProfile.SetBio("I am a developer")

	fmt.Printf("设置Profile.Bio后: User.IsDirty=%v, User.DirtyCount=%d\n", user.IsDirty(), user.GetDirtyFieldCount())
	fmt.Printf("Profile.IsDirty=%v, Profile.DirtyCount=%d\n", userProfile.IsDirty(), userProfile.GetDirtyFieldCount())

	// 重置脏标记
	user.ResetDirty()
	fmt.Printf("重置脏标记后: IsDirty=%v, DirtyCount=%d\n", user.IsDirty(), user.GetDirtyFieldCount())

	// 演示位图的内存优势
	fmt.Println("\n=== 位图内存使用比较 ===")
	fmt.Printf("旧方案(每个字段1个bool): %d bytes\n", 5*1) // 5个字段 * 1 byte per bool
	fmt.Printf("新方案(1个uint64位图): %d bytes\n", 8)   // 1个uint64 = 8 bytes
	fmt.Printf("内存节省: %d bytes\n", 5*1-8)

	fmt.Println("\n=== 测试完成 ===")
}
