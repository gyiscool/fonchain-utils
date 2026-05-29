package translation

import "testing"

func TestGetConvertString(t *testing.T) {

	NewTranslation("", "", "")

	type args struct {
		msg  string
		lang string
		sys  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{args: args{msg: "数据不存在", lang: "EN", sys: "tt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConvertString(tt.args.msg, tt.args.lang, tt.args.sys)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConvertString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetConvertString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
