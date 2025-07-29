package flags

import "github.com/spf13/pflag"

func StringVarPtr(fs *pflag.FlagSet, ptr **string, name, value, usage string) {
	if *ptr == nil {
		*ptr = new(string)
	}
	fs.StringVar(*ptr, name, value, usage)
}

func Int64VarPtr(fs *pflag.FlagSet, ptr **int64, name string, value int64, usage string) {
	if *ptr == nil {
		*ptr = new(int64)
	}
	fs.Int64Var(*ptr, name, value, usage)
}

func BoolVarPtr(fs *pflag.FlagSet, ptr **bool, name string, value bool, usage string) {
	if *ptr == nil {
		*ptr = new(bool)
	}
	fs.BoolVar(*ptr, name, value, usage)
}

func Float64VarPtr(fs *pflag.FlagSet, ptr **float64, name string, value float64, usage string) {
	if *ptr == nil {
		*ptr = new(float64)
	}
	fs.Float64Var(*ptr, name, value, usage)
}

func StringVar(fs *pflag.FlagSet, ptr *string, name, value, usage string) {
	fs.StringVar(ptr, name, value, usage)
}

func StringArrayVar(fs *pflag.FlagSet, ptr *[]string, name string, value []string, usage string) {
	fs.StringArrayVar(ptr, name, value, usage)
}

func Int64Var(fs *pflag.FlagSet, ptr *int64, name string, value int64, usage string) {
	fs.Int64Var(ptr, name, value, usage)
}

func BoolVar(fs *pflag.FlagSet, ptr *bool, name string, value bool, usage string) {
	fs.BoolVar(ptr, name, value, usage)
}

func BoolVarP(fs *pflag.FlagSet, ptr *bool, name, short string, value bool, usage string) {
	fs.BoolVarP(ptr, name, short, value, usage)
}

func Float64Var(fs *pflag.FlagSet, ptr *float64, name string, value float64, usage string) {
	fs.Float64Var(ptr, name, value, usage)
}

func StringSliceVar(fs *pflag.FlagSet, ptr *[]string, name string, value []string, usage string) {
	fs.StringSliceVar(ptr, name, value, usage)
}
