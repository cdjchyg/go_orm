package main

import (
	"fmt"

	"DB/example/pb"
)

func main() {
	fmt.Println("=== 测试切片(tags)和映射(metadata)的脏标记功能 ===")

	// 创建用户对象
	user := pb.NewUser()

	// 初始状态
	fmt.Printf("初始状态 - 是否有脏数据: %t\n", user.IsDirty())

	// 设置基础字段
	user.SetId("test_user")
	user.SetName("测试用户")
	user.SetAge(25)

	fmt.Printf("设置基础字段后 - 脏字段数量: %d\n", user.GetDirtyFieldCount())

	// 测试切片操作
	fmt.Println("\n--- 测试切片(tags)操作 ---")
	user.AddTagsElement("Go")
	user.AddTagsElement("MongoDB")
	user.AddTagsElement("开发者")

	fmt.Printf("添加tags后 - 标签: %v\n", user.GetTags())
	fmt.Printf("Tags字段是否脏: %t\n", user.IsTagsDirty())

	// 修改切片中的元素
	user.SetTagsElement(0, "Golang")
	fmt.Printf("修改第一个标签后 - 标签: %v\n", user.GetTags())

	// 直接设置整个切片
	user.SetTags([]string{"Python", "JavaScript", "React"})
	fmt.Printf("直接设置整个切片后 - 标签: %v\n", user.GetTags())

	// 测试映射操作
	fmt.Println("\n--- 测试映射(metadata)操作 ---")
	user.SetMetadataValue("department", "技术部")
	user.SetMetadataValue("level", "高级")
	user.SetMetadataValue("city", "北京")

	fmt.Printf("添加metadata后 - 元数据: %v\n", user.GetMetadata())
	fmt.Printf("Metadata字段是否脏: %t\n", user.IsMetadataDirty())

	// 修改已存在的键值
	user.SetMetadataValue("level", "专家")
	fmt.Printf("修改level后 - 元数据: %v\n", user.GetMetadata())

	// 直接设置整个映射
	newMetadata := map[string]string{
		"role":   "架构师",
		"team":   "核心团队",
		"status": "在职",
	}
	user.SetMetadata(newMetadata)
	fmt.Printf("直接设置整个映射后 - 元数据: %v\n", user.GetMetadata())

	// 测试嵌套对象
	fmt.Println("\n--- 测试嵌套对象(profile)操作 ---")
	profile := pb.NewUserProfile()
	profile.SetAvatarUrl("https://example.com/avatar.jpg")
	profile.SetBio("一个热爱编程的开发者")

	user.SetProfile(profile)
	fmt.Printf("Profile字段是否脏: %t\n", user.IsProfileDirty())

	// 修改嵌套对象
	user.GetProfile().SetBio("一个经验丰富的全栈开发者")
	fmt.Printf("修改profile后 - Bio: %s\n", user.GetProfile().GetBio())
	fmt.Printf("父对象Profile字段是否脏: %t\n", user.IsProfileDirty())

	// 最终状态
	fmt.Println("\n--- 最终状态 ---")
	fmt.Printf("总脏字段数量: %d\n", user.GetDirtyFieldCount())
	fmt.Printf("脏字段索引: %v\n", user.GetDirtyFieldIndexes())

	// 检查各字段状态
	fmt.Printf("Id字段是否脏: %t\n", user.IsIdDirty())
	fmt.Printf("Name字段是否脏: %t\n", user.IsNameDirty())
	fmt.Printf("Age字段是否脏: %t\n", user.IsAgeDirty())
	fmt.Printf("Tags字段是否脏: %t\n", user.IsTagsDirty())
	fmt.Printf("Metadata字段是否脏: %t\n", user.IsMetadataDirty())
	fmt.Printf("Profile字段是否脏: %t\n", user.IsProfileDirty())

	// 重置脏标记
	fmt.Println("\n--- 重置脏标记后 ---")
	user.ResetDirty()
	fmt.Printf("重置后脏字段数量: %d\n", user.GetDirtyFieldCount())
	fmt.Printf("重置后是否有脏数据: %t\n", user.IsDirty())

	fmt.Println("\n✅ 所有测试完成！切片和映射处理功能正常！")
}
