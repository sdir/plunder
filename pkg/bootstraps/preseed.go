package bootstraps

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

// Preseed const, this is the basis for the configuration that will be modified per use-case
const preseed = `
# Force debconf priority to critical.
debconf debconf/priority select critical
# Override default frontend to Noninteractive
debconf debconf/frontend select Noninteractive

# Preseeding only locale sets language, country and locale.
d-i debian-installer/locale string en_US

# Disable automatic (interactive) keymap detection.
d-i console-setup/ask_detect boolean false
d-i keyboard-configuration/layoutcode string us

### Clock and time zone setup
d-i clock-setup/utc boolean true
d-i time/zone string Europe/GMT
d-i clock-setup/ntp boolean true
d-i clock-setup/ntp-server string 1.pl.pool.ntp.org

### Preseed Early
d-i preseed/early_command string kill-all-dhcp; netcfg
`

const preseedNet = `
### Network configuration
d-i netcfg/wireless_wep string

# Set network interface or 'auto'
d-i netcfg/choose_interface select auto

# Any hostname and domain names assigned from dhcp take precedence over
d-i netcfg/get_gateway string %s
d-i netcfg/get_ipaddress string %s
d-i netcfg/get_nameservers string %s
d-i netcfg/get_netmask string %s
d-i netcfg/use_dhcp string
d-i netcfg/disable_dhcp boolean true

d-i netcfg/get_hostname string ubuntu
d-i netcfg/get_domain string internal

d-i netcfg/hostname string %s`

const preseedDisk = `
### Partitions
d-i partman/mount_style select label

### Boot loader installation
d-i grub-installer/only_debian boolean true
d-i grub-installer/with_other_os boolean true

### Finishing up the installation
d-i finish-install/reboot_in_progress note
d-i cdrom-detect/eject boolean true

### Preseeding other packages
popularity-contest popularity-contest/participate boolean false

### GRUB
grub-pc grub-pc/hidden_timeout  boolean true
grub-pc grub-pc/timeout string  0

### Regular, primary partitions
d-i partman-auto/disk string /dev/sda

# The presently available methods are:
# - regular: use the usual partition types for your architecture
# - lvm:     use LVM to partition the disk
# - crypto:  use LVM within an encrypted partition
d-i partman-auto/method string regular

# You can choose one of the three predefined partitioning recipes:
# - atomic: all files in one partition
# - home:   separate /home partition
# - multi:  separate /home, /usr, /var, and /tmp partitions
d-i partman-auto/choose_recipe select atomic
d-i partman/default_filesystem string ext4

# This makes partman automatically partition without confirmation, provided
# that you told it what to do using one of the methods above.
d-i partman-partitioning/confirm_write_new_label boolean true
d-i partman/choose_partition select finish
d-i partman/confirm boolean true
d-i partman/confirm_nooverwrite boolean true`

const preseedUsers = `
### Account setup
d-i passwd/root-login boolean false
d-i passwd/make-user boolean true
d-i passwd/user-fullname string ubuntu
d-i passwd/username string ubuntu
# TODO probably you need some decent password
d-i passwd/user-password password ubuntu
d-i passwd/user-password-again password ubuntu
d-i user-setup/allow-password-weak boolean true
d-i user-setup/encrypt-home boolean false
`

const preseedPkg = `
### Apt setup
d-i apt-setup/restricted boolean true
d-i apt-setup/universe boolean true
d-i mirror/http/hostname string %s
d-i mirror/http/directory string %s
d-i mirror/country string manual
d-i mirror/http/proxy string

### Base system installation
d-i base-installer/install-recommends boolean false

### Package selection
tasksel tasksel/first multiselect
tasksel/skip-tasks multiselect server
d-i pkgsel/ubuntu-standard boolean false

# Allowed values: none, safe-upgrade, full-upgrade
d-i pkgsel/upgrade select none
d-i pkgsel/ignore-incomplete-language-support boolean true
d-i pkgsel/include string openssh-server

# Language pack selection
d-i pkgsel/install-language-support boolean false
d-i pkgsel/language-pack-patterns string
d-i pkgsel/language-packs multiselect
# or ...
#d-i pkgsel/language-packs multiselect en, pl

# Policy for applying updates. May be "none" (no automatic updates),
# "unattended-upgrades" (install security updates automatically), or
# "landscape" (manage system with Landscape).
d-i pkgsel/update-policy select unattended-upgrades
d-i pkgsel/updatedb boolean false
`

const preseedCmd = `
d-i preseed/late_command string \
    in-target sed -i 's/^%sudo.*$/%sudo ALL=(ALL:ALL) NOPASSWD: ALL/g' /etc/sudoers; \
    in-target /bin/sh -c "echo 'Defaults env_keep += \"SSH_AUTH_SOCK\" >> /etc/sudoers"; \
    in-target mkdir -p /home/ubuntu/.ssh; \
    in-target /bin/sh -c "echo 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDbT6GcIYdJB96MVL35dhIAy9tRx0Hl/0HnHuk/ep+NRuGTExepHEO/8Dop67MT24e3Q6VXORYPeAHc3sgWrP3D7NrzKJkgE44SSL1v/94BpHJ0yNsver79DS73FU+NOCOJXWoxvB40F5UhAIVOkWqs8dLOugqKfKZfovetu6RvgEDcjR79Ndqk6JBqPotybQ9Kfpgt/wyCponBWXZn4Q+sAQAT1pg6FUICOh4/SZwDv29E7x0/1hD9tO+r2x1ZNo/VMecsecBdWPixXlpS1Az16bmYrgXULTfO8Y9174bu2MlnlPGLmm7wBO4PL7L9WiiG3it82ZDzi7PO59yUIUpL ubuntu insecure public key' >> /home/ubuntu/.ssh/authorized_keys"; \
    in-target chown -R ubuntu:ubuntu /home/ubuntu/; \
	in-target chmod -R go-rwx /home/ubuntu/.ssh/authorized_keys;
`

//BuildPreeSeedConfig - Creates a new presseed configuration using the passed data
func (config *ServerConfig) BuildPreeSeedConfig() string {
	// TODO - this is broken
	if config.SSHKeyPath != "" {
		err := config.ReadKeyFromFile(config.SSHKeyPath)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	parsedNet := fmt.Sprintf(preseedNet, config.Gateway, config.IPAddress, config.NameServer, config.Subnet, config.ServerName)
	parsedPkg := fmt.Sprintf(preseedPkg, config.RepositoryAddress, config.MirrorDirectory)
	return fmt.Sprintf("%s%s%s%s%s%s", preseed, preseedDisk, parsedNet, parsedPkg, preseedUsers, preseedCmd)
}
