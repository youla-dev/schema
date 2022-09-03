package cmd

func subjectName(kafkaTopic, record string) string {
	return kafkaTopic + "-" + record + "-value"
}
