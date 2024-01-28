/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package flag

import (
	kingpin "github.com/alecthomas/kingpin/v2"
	"trellis.tech/common.v2/logger/prometheus"
)

// AddFlags adds the flags used by this package to the Kingpin application.
// To use the default Kingpin application, call AddFlags(kingpin.CommandLine)
func AddFlags(a *kingpin.Application, config *prometheus.Config) {
	a.Flag("log.level", "Only log messages with the given severity or above. One of: [debug, info, warn, error]").
		Default("info").StringVar(&config.Level)
	a.Flag("log.file-name", "Path to the log directory.").Default("").StringVar(&config.FileName)
	a.Flag("log.move-file-type", "Move file type.[0-Never move files, 1-Move files every minute, 2-Move files every hour, 3-Move files every day]").
		Default("3").IntVar(&config.MoveFileType)
	a.Flag("log.max-length", "The size of one file, 0-Undivided file.").
		Default("0").Int64Var(&config.MaxLength)
	a.Flag("log.max-backups", "The maximum number of saved files, 0-Save all files.").
		Default("10").IntVar(&config.MaxBackups)
	a.Flag("log.std_printers", "The standard printers: stderr, stdout").StringsVar(&config.StdPrinters)
	a.Flag("log.time_format", "The time format").Default("").StringVar(&config.TimeFormat)
	a.Flag("log.caller", "caller").Default("false").BoolVar(&config.Caller)
	a.Flag("log.caller_skip", "caller skip").Default("3").IntVar(&config.CallerSkip)
}
