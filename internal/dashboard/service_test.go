package dashboard

import "testing"

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr bool
	}{
		{name: "valid", request: Request{LeagueID: "league", UserID: "user", Week: 1}},
		{name: "missing league", request: Request{UserID: "user", Week: 1}, wantErr: true},
		{name: "missing user", request: Request{LeagueID: "league", Week: 1}, wantErr: true},
		{name: "week zero", request: Request{LeagueID: "league", UserID: "user"}, wantErr: true},
		{name: "week too high", request: Request{LeagueID: "league", UserID: "user", Week: 19}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRequest(test.request)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateRequest() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}
