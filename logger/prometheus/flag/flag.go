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
	"trellis.tech/trellis/common.v2/logger/prometheus"
)

// AddFlags adds the flags used by this package to the Kingpin application.
// To use the default Kingpin application, call AddFlags(kingpin.CommandLine)
func AddFlags(a *kingpin.Application, config *prometheus.Config) {
	a.Flag("log.level", "Only log messages with the given severity or above. One of: [trace, debug, info, warn, error, fatal, panic]").Default("info").StringVar(&config.Level)
	a.Flag("log.file-name", "Path to the log directory.").Default("").StringVar(&config.FileName)
	a.Flag("log.move-file-type", "Move file type.[1-Move files every minute, 2-Move files every hour, 3-Move files every day]").
		Default("3").IntVar(&config.MoveFileType)
	a.Flag("log.max-length", "The size of one file, 0-Undivided file.").
		Default("0").Int64Var(&config.MaxLength)
	a.Flag("log.max-backups", "The maximum number of saved files, 0-Save all files.").
		Default("10").UintVar(&config.MaxBackups)
	a.Flag("log.caller", "caller").Default("false").BoolVar(&config.Caller)
}
