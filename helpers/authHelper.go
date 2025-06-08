package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(c*gin.Context,role string) (err error){
	userType :=c.GetString("user_type")
	err=nil
	if userType!=role {
		err = errors.New ("Unauthorized to access this resource")
	}
	return err
}


func MatchUserTypeToUid(c *gin.Context,userId string) (err error){
	//Hm sb authenticate step me extract kr lete hai tokens se to yha sirf get string se sb mil skta hai
	//yeh JWT token me se usertype nikal rha hai hme kch sent krne ki need nhi hai 
	userType := c.GetString("user_type")
	uid:=c.GetString("uid")
	err=nil
	
	if userType == "USER" && uid!=userId{
		err =errors.New("Unauthorized to access this resource")

		return err
	}
	err= CheckUserType(c,userType)
	return err
}