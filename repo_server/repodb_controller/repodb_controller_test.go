package repodbcontroller

import (
	"reflect"
	"testing"
)

func packageEqual(a, b Package) bool {
	return a.Name == b.Name &&
		a.Version == b.Version &&
		a.PackageName == b.PackageName &&
		a.Status == b.Status
}

func packagesEqual(a, b []Package) bool {
	if len(a) != len(b) {
		return false
	}

	for i, val := range a {
		if !packageEqual(val, b[i]) {
			return false
		}
	}
	return true
}

func TestNewRepoDBController(t *testing.T) {
	tests := []struct {
		name string
		want *RepoDBController
	}{
		{
			name: "Test New",
			want: &RepoDBController{
				host: "localhost",
				user: "admin",
				pwd:  "admin",
				port: "27017",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRepoDBController(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRepoDBController() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepoDBController_ConnectDB(t *testing.T) {
	tests := []struct {
		name    string
		r       *RepoDBController
		wantErr bool
	}{
		{
			name: "Conn Success",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.ConnectDB(); (err != nil) != tt.wantErr {
				t.Errorf("RepoDBController.ConnectDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepoDBController_GetPackagesByName(t *testing.T) {
	type args struct {
		repoNames []string
		name      string
	}
	tests := []struct {
		name    string
		r       *RepoDBController
		args    args
		want    []Package
		wantErr bool
	}{
		{
			name: "Has_all_repos_has_full_pkg_name_matching",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				repoNames: []string{
					"test",
				},
				name: "Placeholder",
			},
			want: []Package{
				Package{
					Name:        "Placeholder",
					Version:     "v1.0.0",
					PackageName: "ph",
					Status:      "online",
				},
			},
			wantErr: false,
		},
		{
			name: "Has_all_repos_has_partial_pkg_name_matching",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				repoNames: []string{
					"test",
				},
				name: "holder",
			},
			want: []Package{
				Package{
					Name:        "Cupholder",
					Version:     "v2.0.0",
					PackageName: "ch",
					Status:      "online",
				},
				Package{
					Name:        "Placeholder",
					Version:     "v1.0.0",
					PackageName: "ph",
					Status:      "online",
				},
			},
			wantErr: false,
		},
		{
			name: "Has_all_repos_no_pkg_name_matching",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				repoNames: []string{
					"test",
				},
				name: "Nothing",
			},
			want:    []Package{},
			wantErr: false,
		},
		{
			name: "Has_some_repos_has_partial_pkg_name_matching",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				repoNames: []string{
					"test",
					"norepo",
				},
				name: "holder",
			},
			want: []Package{
				Package{
					Name:        "Cupholder",
					Version:     "v2.0.0",
					PackageName: "ch",
					Status:      "online",
				},
				Package{
					Name:        "Placeholder",
					Version:     "v1.0.0",
					PackageName: "ph",
					Status:      "online",
				},
			},
			wantErr: false,
		},
		{
			name: "no_repo",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				repoNames: []string{
					"norepo",
				},
				name: "holder",
			},
			want:    []Package{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.ConnectDB()
			got, err := tt.r.GetPackagesByName(tt.args.repoNames, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("RepoDBController.GetPackagesByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !packagesEqual(got, tt.want) {
				t.Errorf("RepoDBController.GetPackagesByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepoDBController_GetPackageByPkgName(t *testing.T) {
	type args struct {
		pkgName  string
		repoName string
	}
	tests := []struct {
		name    string
		r       *RepoDBController
		args    args
		want    Package
		wantErr bool
	}{
		{
			name: "has_repo_has_package",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				pkgName:  "ph",
				repoName: "test",
			},
			want: Package{
				Name:        "Placeholder",
				Version:     "v1.0.0",
				PackageName: "ph",
				Status:      "online",
			},
			wantErr: false,
		},
		{
			name: "has_repo_no_package",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				pkgName:  "nopkg",
				repoName: "test",
			},
			want:    Package{},
			wantErr: false,
		},
		{
			name: "no_repo",
			r: &RepoDBController{
				host: "localhost",
				user: "repo",
				pwd:  "reporepo",
				port: "27017",
			},
			args: args{
				pkgName:  "ph",
				repoName: "norepo",
			},
			want:    Package{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.ConnectDB()
			got, err := tt.r.GetPackageByPkgName(tt.args.pkgName, tt.args.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RepoDBController.GetPackageByPkgName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !packageEqual(got, tt.want) {
				t.Errorf("RepoDBController.GetPackageByPkgName() = %v, want %v", got, tt.want)
			}
		})
	}
}
