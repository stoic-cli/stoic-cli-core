package tool

type Runner interface {
	Setup(checkout Checkout) error
	Run(checkout Checkout, name string, args []string) error
}
