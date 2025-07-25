package main

import (
	"DB/example/pb"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 连接MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("testdb").Collection("users")

	// 创建用户对象
	user := pb.NewUser()

	// 设置用户信息
	user.SetId("user_001")
	user.SetName("张三")
	user.SetEmail("zhangsan@example.com")
	user.SetAge(25)

	// 添加标签
	user.AddTagsElement("开发者")
	user.AddTagsElement("Go")
	user.AddTagsElement("MongoDB")

	// 设置元数据
	user.SetMetadataValue("department", "技术部")
	user.SetMetadataValue("level", "高级")

	// 创建用户详情
	profile := pb.NewUserProfile()
	profile.SetAvatarUrl("https://example.com/avatar.jpg")
	profile.SetBio("一个热爱编程的开发者")

	user.SetProfile(profile)

	// 打印初始状态
	fmt.Println("=== 初始创建后的状态 ===")
	fmt.Printf("用户ID: %s\n", user.GetId())
	fmt.Printf("用户名: %s\n", user.GetName())
	fmt.Printf("邮箱: %s\n", user.GetEmail())
	fmt.Printf("年龄: %d\n", user.GetAge())
	fmt.Printf("标签: %v\n", user.GetTags())
	fmt.Printf("元数据: %v\n", user.GetMetadata())
	fmt.Printf("是否有脏数据: %t\n", user.IsDirty())
	fmt.Printf("脏字段数量: %d\n", user.GetDirtyFieldCount())

	// 保存到MongoDB（简化版）
	userDoc := bson.M{
		"id":       user.GetId(),
		"name":     user.GetName(),
		"email":    user.GetEmail(),
		"age":      user.GetAge(),
		"tags":     user.GetTags(),
		"metadata": user.GetMetadata(),
	}

	filter := bson.M{"id": user.GetId()}
	opts := options.Replace().SetUpsert(true)
	_, err = collection.ReplaceOne(context.TODO(), filter, userDoc, opts)
	if err != nil {
		log.Fatal("保存失败:", err)
	}
	fmt.Println("✅ 用户创建并保存成功")

	// 模拟一些修改
	fmt.Println("\n=== 模拟修改操作 ===")
	user.SetAge(26)                      // 年龄+1
	user.AddTagsElement("数据库")           // 添加新标签
	user.SetMetadataValue("level", "专家") // 修改级别

	// 只修改profile中的一个字段
	user.GetProfile().SetBio("一个经验丰富的全栈开发者")

	// 打印修改后的状态
	fmt.Println("=== 修改后的状态 ===")
	fmt.Printf("年龄: %d\n", user.GetAge())
	fmt.Printf("标签: %v\n", user.GetTags())
	fmt.Printf("元数据: %v\n", user.GetMetadata())
	fmt.Printf("是否有脏数据: %t\n", user.IsDirty())
	fmt.Printf("脏字段数量: %d\n", user.GetDirtyFieldCount())
	fmt.Printf("脏字段索引: %v\n", user.GetDirtyFieldIndexes())

	// 检查具体哪些字段是脏的
	fmt.Printf("Age字段是脏的: %t\n", user.IsAgeDirty())
	fmt.Printf("Tags字段是脏的: %t\n", user.IsTagsDirty())
	fmt.Printf("Metadata字段是脏的: %t\n", user.IsMetadataDirty())
	fmt.Printf("Profile字段是脏的: %t\n", user.IsProfileDirty())

	// 再次保存
	userDoc = bson.M{
		"id":       user.GetId(),
		"name":     user.GetName(),
		"email":    user.GetEmail(),
		"age":      user.GetAge(),
		"tags":     user.GetTags(),
		"metadata": user.GetMetadata(),
	}
	_, err = collection.ReplaceOne(context.TODO(), filter, userDoc, opts)
	if err != nil {
		log.Fatal("更新失败:", err)
	}
	fmt.Println("✅ 用户更新成功")

	// 重置脏标记
	user.ResetDirty()
	fmt.Printf("重置后是否有脏数据: %t\n", user.IsDirty())

	// 演示批量操作
	fmt.Println("\n=== 批量操作演示 ===")
	var users []*pb.User

	// 创建多个用户
	for i := 0; i < 3; i++ {
		u := pb.NewUser()
		u.SetId(fmt.Sprintf("user_%03d", i+2))
		u.SetName(fmt.Sprintf("用户%d", i+2))
		u.SetEmail(fmt.Sprintf("user%d@example.com", i+2))
		u.SetAge(int32(20 + i))
		u.AddTagsElement("测试用户")
		users = append(users, u)
	}

	// 简化版批量保存
	var docs []interface{}
	for _, u := range users {
		doc := bson.M{
			"id":    u.GetId(),
			"name":  u.GetName(),
			"email": u.GetEmail(),
			"age":   u.GetAge(),
			"tags":  u.GetTags(),
		}
		docs = append(docs, doc)
	}

	_, err = collection.InsertMany(context.TODO(), docs)
	if err != nil {
		log.Fatal("批量保存失败:", err)
	}
	fmt.Println("✅ 批量保存成功")

	// 验证保存结果
	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("数据库中总共有 %d 个用户\n", count)
}
