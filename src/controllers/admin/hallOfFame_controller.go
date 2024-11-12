package controller_admin

import (
	"LearnGo/helper"
	"LearnGo/models"
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func AvgStudentScores(semester string, course_id bson.ObjectID) []avgStudentScore {

	coursesCollection := models.CourseModel()

	// Tìm khóa học với course_id
	var course models.InterfaceCourse
	err := coursesCollection.FindOne(context.TODO(), bson.M{"_id": course_id}).Decode(&course)
	if err != nil {
		log.Fatal(err)
	}
	HS := course.HS
	// Tìm danh sách điểm của sinh viên trong học kỳ và khóa học cụ thể
	scoresCollection := models.ResultScoreModel()
	cursor, err := scoresCollection.Find(context.TODO(), bson.M{"course_id": course_id, "semester": semester})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var resulScores []models.InterfaceResultScore
	if err = cursor.All(context.TODO(), &resulScores); err != nil {
		log.Fatal(err)
	}
	// Khởi tạo và gán giá trị cho slice avgScores trên cùng một dòng
	totalSize := 0
	for _, result := range resulScores {
		totalSize += len(result.SCORE)
	}
	i := 0
	avgScores := make([]avgStudentScore, totalSize)
	for _, result := range resulScores {
		for _, score := range result.SCORE {
			avgScores[i].MSSV = score.MSSV
			avgScores[i].AvgScore = helper.AvgScore(score.Data, HS[:])
			i++
		}
	}
	return avgScores
}

func MergeSort(avgScores []avgStudentScore) []avgStudentScore {
	if len(avgScores) <= 1 {
		return avgScores
	}

	mid := len(avgScores) / 2
	left := MergeSort(avgScores[:mid])
	right := MergeSort(avgScores[mid:])

	return merge(left, right)
}

func merge(left, right []avgStudentScore) []avgStudentScore {
	var result []avgStudentScore
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i].AvgScore > right[j].AvgScore { // Sắp xếp giảm dần
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	// Thêm các phần tử còn lại
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

func CheckDuplicateHOF(collection *mongo.Collection, semester string, course_id bson.ObjectID) bool {

	filter := bson.M{
		"semester":  semester,
		"course_id": course_id,
	}

	//kiểm tra xem có bản ghi nào không
	var result bson.M
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return false // Không tìm thấy bản ghi
	} else if err != nil {
		return false // Có lỗi khác
	}

	return true
}

func CreateHallOfFame(c *gin.Context) {
	scoresCollection := models.ResultScoreModel()
	var results []models.InterfaceResultScore
	cursor, err := scoresCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	processed := make(map[string]bool)
	collection := models.HallOfFameModel()
	for _, result := range results {
		key := result.Semester + "-" + result.CourseID.Hex()
		if _, found := processed[key]; !found {
			processed[key] = true
			avgStudentScores := AvgStudentScores(result.Semester, result.CourseID)
			studentHOF := MergeSort(avgStudentScores)
			var data bson.A
			length := min(10, len(studentHOF))
			for i := 0; i < length; i++ {
				student := studentHOF[i]
				data = append(data, bson.M{"mssv": student.MSSV, "dtb": student.AvgScore})
			}
			if !CheckDuplicateHOF(collection, result.Semester, result.CourseID) {

				collection.InsertOne(context.TODO(), bson.M{
					"semester":  result.Semester,
					"course_id": result.CourseID,
					"data":      data,
				})
			} else {
				filter := bson.M{
					"semester":  result.Semester,
					"course_id": result.CourseID,
				}
				update := bson.M{
					"$set": bson.M{
						"data": data},
				}
				collection.UpdateOne(context.TODO(), filter, update)

			}
		}
	}
	c.JSON(200, gin.H{
		"code":    "success",
		"message": "Cập nhật Hall Of Fame thành công",
	})
}
