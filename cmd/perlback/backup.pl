#!/usr/bin/perl -w


# Include libraries
use strict;
use warnings;
use POSIX 'strftime';
use File::Basename;
use File::Copy;
use File::Path qw(make_path remove_tree);
use File::Find;
use Term::ANSIColor;
use File::Basename;
use Getopt::Long;


# Initialize
my $version = "1.1.1706.10101";
my $now = time();
my $date = strftime '%Y%m%d', localtime( $now );
my $diff_sec = 24 * 60 * 60 * 1; # 1 Day
my $tmpdir = "/tmp/";
my $old_data_file = $tmpdir . "backup_hash.old";
my %new_data_sheet = ();
my @modified = ();


# Check arguments
my $src_dir;
my $dst_dir;
my $verbose;
my $report_file;
GetOptions (
    's=s' => \$src_dir,
    'd=s' => \$dst_dir,
    'v'   => \$verbose,
    "r=s" => \$report_file
);
if ( ! $src_dir || ! $dst_dir ) {
    print "Description : File Incremental Backup Manager $version", "\n";
    print "Usage       : backup.pl -s[=SOURCE] -d[=DEST] [-v, verbose] [-r[=report_filepath]]", "\n";
    print "Example     : backup.pl -s=/backup/current -d=/backup/20170601", "\n";
    print "Example     : backup.pl -s=/backup/current -d=/backup/20170601 -v", "\n";
    print "Example     : backup.pl -s=/backup/current -d=/backup/20170601 -r=/home/report/report.log", "\n";
    print "Daily Job   : 5 0 * * * /root/backup.pl -s=/backup/current -d=/backup/\$(date +\\%Y\\%m\\%d) /dev/null 2>&1", "\n";
    exit;
}


# Check if source directory exists
$src_dir =~ s#/*$#/#;
unless ( -e $src_dir and -d $src_dir ) {
    die( $src_dir . " does not exist\n" );
}


# Set destination directory
$dst_dir =~ s#/*$#/#;
if ( ! -e $dst_dir ) {
    mkdir $dst_dir, 0755 or die( "Cannot create $dst_dir\n");
}


# Set path of report log file
if ( ! $report_file ) {
    $report_file = "/tmp/report.log";
}
die( dirname( $report_file ) . " does not exist\n" ) if ( ! -e dirname( $report_file ) );
die( $report_file . " is not a file\n" ) if ( -e $report_file and ! -f $report_file );


# Run
find( \&do_differential_backup, ( $src_dir ) );
find( \&touch_all_directories, ( $src_dir ) );
generate_report();
write_new_data_sheet_into_file();


# Functions
sub write_new_data_sheet_into_file {
    unlink $old_data_file if ( -e $old_data_file );
    open( WRITE, ">>", $old_data_file );
    foreach my $f ( keys %new_data_sheet ) {
        print WRITE $f, "\n";
    }
    close( WRITE );
}


sub generate_report {
    my $time_str = localtime( $now );
	my %old_data_sheet = read_file( "/tmp/backup_hash.old" );
    open( REPORT, ">>", $report_file );
	my @diff;

    # Write deleted items
	@diff = diff( \%old_data_sheet, \%new_data_sheet );
	foreach ( @diff ) {
        if ( $verbose ) {
            print_ansi( "red", "[Deleted] " );
            print $_, "\n";
        }
        printf( REPORT "%s\t[%s]\t%s\n",  $time_str, "deleted", $_ );
	}

    # Write added items
	@diff = diff( \%new_data_sheet, \%old_data_sheet );
	foreach ( @diff ) {
        if ( $verbose ) {
            print_ansi( "yellow", "[Added] " );
		    print  $_, "\n";
        }
        my $mtime = ( stat ( $_ ) )[9];
        my $mtime_str = localtime( $mtime );
        printf( REPORT "%s\t[%s]\t%s\n",  $mtime_str, "added", $_ );
	}

    # Write modified items
    foreach ( @modified ) {
        next if ( ! $old_data_sheet{ $_  } );
        if ( $verbose ) {
            print_ansi( "green", "[Modified] " );
            print $_, "\n";
        }

        my $mtime = ( stat ( $_ ) )[9];
        my $mtime_str = localtime( $mtime );
        printf( REPORT "%s\t[%s]\t%s\n",  $mtime_str, "modified", $_ );
    }

    close( REPORT );
}




# Functions
sub do_differential_backup {

    if ( -f $File::Find::name ) {

        # Get the last modified time of the file
        my $mtime = ( stat ( $File::Find::name ) )[9];

        # Check time difference
        if ($now - $mtime < $diff_sec) {
            my ( $filename, $dirs, $suffix ) = fileparse( $File::Find::name );

            # Make directory
            $dirs =~ s|^$src_dir|$dst_dir|;
            if ( ! -d $dirs ) {
                make_path($dirs);
            }

            # Copy file
            copy( $File::Find::name, $dirs . $filename ) or die "Copy failed: $!";
            utime $mtime, $mtime, $dirs . $filename;
            push @modified, $File::Find::name;
        }

        # Fill new data sheet
        $new_data_sheet{ $File::Find::name } = 1;

    } else {
        $new_data_sheet{ $File::Find::name . "/" } = 1;
    }
}



sub print_ansi {
    my ( $color, $str ) = @_;
    print color( "bold $color" );
    print $str;
    print color( "reset" );
}


# Set modified time of destination directory to source directory's it
sub touch_all_directories {
    if ( -d $File::Find::name ) {

        my $dir1 = $File::Find::name;
        my $dir2 = $dir1;
        $dir2 =~ s|^$src_dir|$dst_dir|;

        # Get the last modified time of the file
        my $mtime = ( stat ( $dir1 ) )[9];
        utime $mtime, $mtime, $dir2;
    }
}


# Read a file and save it into hash table
sub read_file {
    my $fp = shift;
    my %hash = ();

    if ( -e $fp ) {
        open( READ, "<", $fp );
        my @row = <READ>;
        chomp( @row );
        %hash = map { $_ => 1 } @row;
        close( READ );
    }

    return %hash;
}


sub diff {
    my ( $h1, $h2 ) = @_;
    my @diff = ();
	foreach ( keys %$h1 ) {
		push( @diff, $_ ) unless exists $$h2{ $_ };
	}
    return @diff;
}
