package wordwrap_test

import (
	"os"

	"github.com/mdigger/wordwrap"
)

func Example() {
	// source unwrapped text
	source := `Lorem ipsum dolor sit amet, lectus sed ut at lacinia. ` +
		`A adipiscing. Vel placerat, ornare vel consectetur integer. Et ` +
		`molestie ante mauris, sociis aliqua senectus et. Risus wisi ` +
		`fringilla mauris massa vestibulum, ante est, quis euismod ac ` +
		`suspendisse, sem sodales ligula eleifend tincidunt, nemo donec ` +
		`porta viverra. Volutpat hymenaeos eu non neque sint. Torquent ` +
		`mauris ante et, suspendisse aliquam nunc, urna sem a ornare sed ` +
		`ante laoreet.`
	w := wordwrap.New(os.Stdout, 50) // init wrap writer
	prefix := "> "                   // define a prefix
	w.SetPrefix(prefix)              // set prefix for new lines
	w.WriteString(prefix)            // add prefix to first line
	w.WriteString(source)            // write other text
	// Output:
	// > Lorem ipsum dolor sit amet, lectus sed ut at
	// > lacinia. A adipiscing. Vel placerat, ornare vel
	// > consectetur integer. Et molestie ante mauris,
	// > sociis aliqua senectus et. Risus wisi fringilla
	// > mauris massa vestibulum, ante est, quis euismod
	// > ac suspendisse, sem sodales ligula eleifend
	// > tincidunt, nemo donec porta viverra. Volutpat
	// > hymenaeos eu non neque sint. Torquent mauris
	// > ante et, suspendisse aliquam nunc, urna sem a
	// > ornare sed ante laoreet.
}
