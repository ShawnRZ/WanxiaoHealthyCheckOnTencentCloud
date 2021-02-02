module WanxiaoHealthyCheckOnTencentCloud

go 1.15

require (
	github.com/FNDHSTD/logor v0.0.0-20210128050834-84504dfb2410
	github.com/tencentyun/scf-go-lib v0.0.0-20200624065115-ba679e2ec9c9 // direct
	report v0.0.0
)

replace report => ./report
