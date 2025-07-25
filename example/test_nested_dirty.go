package main

import (
	"fmt"

	"DB/example/pb"
)

func main() {
	fmt.Println("=== 测试嵌套脏标记同步机制 ===")

	// 创建用户对象
	user := pb.NewUser()
	user.SetId("user_001")
	user.SetName("张三")

	// 创建profile对象
	profile := pb.NewUserProfile()
	profile.SetAvatarUrl("https://example.com/avatar.jpg")
	profile.SetBio("原始简介")

	// 设置profile到user
	user.SetProfile(profile)

	// 重置所有脏标记，模拟保存后的状态
	user.ResetDirty()
	profile.ResetDirty()

	fmt.Println("=== 重置脏标记后的状态 ===")
	fmt.Printf("User.IsDirty: %v\n", user.IsDirty())
	fmt.Printf("User.IsProfileDirty: %v\n", user.IsProfileDirty())
	fmt.Printf("Profile.IsDirty: %v\n", profile.IsDirty())
	fmt.Printf("Profile.IsBioDirty: %v\n", profile.IsBioDirty())

	fmt.Println("\n=== 通过user.GetProfile().SetBio()修改profile中的bio字段 ===")

	// 关键测试：通过GetProfile()获取的对象修改字段
	user.GetProfile().SetBio("更新后的简介")

	fmt.Println("修改后的状态:")
	fmt.Printf("User.IsDirty: %v (应该为true)\n", user.IsDirty())
	fmt.Printf("User.IsProfileDirty: %v (应该为true)\n", user.IsProfileDirty())
	fmt.Printf("Profile.IsDirty: %v (应该为true)\n", user.GetProfile().IsDirty())
	fmt.Printf("Profile.IsBioDirty: %v (应该为true)\n", user.GetProfile().IsBioDirty())

	fmt.Println("\n=== 直接修改profile对象中的其他字段 ===")
	user.GetProfile().SetAvatarUrl("https://example.com/new_avatar.jpg")

	fmt.Println("再次修改后的状态:")
	fmt.Printf("User.IsDirty: %v (应该为true)\n", user.IsDirty())
	fmt.Printf("User.IsProfileDirty: %v (应该为true)\n", user.IsProfileDirty())
	fmt.Printf("Profile.IsDirty: %v (应该为true)\n", user.GetProfile().IsDirty())
	fmt.Printf("Profile.IsAvatarUrlDirty: %v (应该为true)\n", user.GetProfile().IsAvatarUrlDirty())

	fmt.Println("\n=== 验证数据一致性 ===")
	fmt.Printf("Bio内容: %s\n", user.GetProfile().GetBio())
	fmt.Printf("AvatarUrl内容: %s\n", user.GetProfile().GetAvatarUrl())

	fmt.Println("\n=== 测试完成 ===")
	if user.IsDirty() && user.IsProfileDirty() && user.GetProfile().IsDirty() {
		fmt.Println("✅ 嵌套脏标记同步机制工作正常！")
	} else {
		fmt.Println("❌ 嵌套脏标记同步机制存在问题")
	}
}
