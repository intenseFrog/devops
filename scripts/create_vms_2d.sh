#!/bin/bash
#apt install qemu libvirt-bin virtinit libguestfs-tools

TMPDIR="/tmp/vmtmp"
MNTDIR="/devops/qcow2"
BASEIMG="$6"

pre() {
  hostname=$1
  test -d $MNTDIR || mkdir -p $MNTDIR

  #tmpdir
  if [ ! -d $TMPDIR ];then
    mkdir -p $TMPDIR
  else
    files1=$(ls $TMPDIR)
    if [ ! -z "$files1" ];then
      umount $TMPDIR
      files2=$(ls $TMPDIR)
      if [ ! -z "$files2" ];then
        echo "Directory has been perent and Some files in $TMPDIR"
        exit 1;
      fi
    fi
  fi

  ##libvirtd
  #libvirt_file1="/etc/libvirt/qemu/networks/default.xml"
  #libvirt_file2="/etc/libvirt/qemu/networks/autostart/default.xml"
  #if [[ -e $libvirt_file1 ]] || [[ -e $libvirt_file2 ]];then
  #  rm -f $libvirt_file1 $libvirt_file12
  #  sed -i "/^#user = \"root\"/s/#//g"  /etc/libvirt/qemu.conf
  #  sed -i "/^#group = \"root\"/s/#//g" /etc/libvirt/qemu.conf
  #  sed -i "/^security_driver/d" /etc/libvirt/qemu.conf
  #  sed -i "/^#security_driver/asecurity_driver=\"none\"" /etc/libvirt/qemu.conf
  #  sed -i "/redhat-release/d" /usr/lib64/guestfs/supermin.d/hostfiles
  #  systemctl restart libvirtd
  #  systemctl enable libvirtd
  #fi

  #check vm
  con="yes"
  for vm in `virsh list --all --name`;do
    if [[ $vm == ${hostname} ]]; then
      echo  -e "!!!!!!!!!!!!!!--vm $vm has been present--!!!!!!!!!!!!!!\n"
      #read -p "!!--\"Y|y|YES|yes\" destroy ${hostname} or \"N|n|NO|no\" exit ::: " con
      case $con in
      [Yy][Ee][Ss]|[Yy])
        virsh destroy ${hostname}
        virsh undefine ${hostname}
        rm -f $MNTDIR/${hostname}.qcow2
        rm -f $MNTDIR/${hostname}-disk-*.qcow2
        if [ -e "/etc/libvirt/qemu/${hostname}.xml" ];then
          rm -f "/etc/libvirt/qemu/${hostname}.xml"
        fi
        ;;
      [Nn][Oo]|[Nn])
        exit 0
        ;;
      *)
        echo "!!!!!!!!!!!!!!!!--Wrong exit--!!!!!!!!!!!!!!"
        exit 1
        ;;
      esac
    else
      continue
    fi
  done
}


function make_net_config_and_return_bridge() {
  device_info=$1
  device_info=${device_info//;/ }
  device_tmpfile="/tmp/tmp_device.txt"
  #rm -f $TMPDIR/etc/sysconfig/network-scripts/ifcfg-*
  if [ -e $device_tmpfile ];then
    rm -rf $device_tmpfile
  fi
  count=0
  cat $TMPDIR/etc/os-release  | grep ^NAME | grep "Ubuntu" > /dev/null
  if [ $? -eq 0 ]; then
    ## ubuntu
    net_config_path="$TMPDIR/etc/network/interfaces"
    for device in ${device_info};do
      _device=$(echo $device | awk -F "#" '{print $1,$2,$3,$4,$5}')
      echo ${_device}|while read eth_br eth_ip eth_nm eth_gw eth_dns;do
        if [ "$eth_br" ];then
          echo "$eth_br" >> $device_tmpfile
          eth_dev="eth$count"
          echo >> $net_config_path
          echo "auto $eth_dev" >> $net_config_path
          if [[ "$eth_ip" != "none" ]] && [[ "$eth_nm" != "none" ]];then
            echo "iface $eth_dev inet static" >> $net_config_path
            echo "address $eth_ip" >> $net_config_path
            echo "netmask $eth_nm" >> $net_config_path
            #if [ "$eth_gw" != "none" ];then
            if [ ! -z "$eth_gw" ];then
              echo "gateway $eth_gw" >> $net_config_path
            fi
            if [ ! -z "$eth_dns" ];then
                echo "dns-nameservers $eth_dns" >> $net_config_path
            fi
          elif [[ "$eth_ip" == "none" ]] && [[ "$eth_nm" == "none" ]];then
            echo "iface $eth_dev inet manual" >> $net_config_path
          fi
        fi
      done
      count=$((count+1))
    done
  else
    ##centos
    for device in ${device_info};do
      net_config_path="$TMPDIR/etc/sysconfig/network-scripts/ifcfg-eth$count"
      if [ ! -f $net_config_path ]; then
        cp $TMPDIR/etc/sysconfig/network-scripts/ifcfg-eth0 $net_config_path
        sed -i "s/^DEVICE.*$/DEVICE=eth$count/g" $net_config_path
      fi
      _device=$(echo $device | awk -F "#" '{print $1,$2,$3,$4,$5}')
      echo ${_device}|while read eth_br eth_ip eth_nm eth_gw eth_dns;do
        if [ "$eth_br" ];then
          echo "$eth_br" >> $device_tmpfile
          eth_dev="eth$count"
          if [[ "$eth_ip" != "none" ]] && [[ "$eth_nm" != "none" ]];then
              sed -i "s/^ONBOOT.*$/ONBOOT=\"yes\"/g" $net_config_path
              sed -i "s/^BOOTPROTO.*$/BOOTPROTO=\"static\"/g" $net_config_path
              sed -i "s/^IPADDR.*$/IPADDR=\"$eth_ip\"/g" $net_config_path
              sed -i "s/^DNS1.*$/DNS1=\"$eth_dns\"/g" $net_config_path
            if [ "$eth_gw" != "none" ];then
              sed -i "s/^GATEWAY.*$/GATEWAY=\"$eth_gw\"/g" $net_config_path
            fi
          elif [[ "$eth_ip" == "none" ]] && [[ "$eth_nm" == "none" ]];then
            sed -i "s/^BOOTPROTO.*$/BOOTPROTO=\"manual\"/g" $net_config_path
          fi
        fi
      done
      count=$((count+1))
    done
  fi
  #echo >> $net_config_path
  #echo "up route add -net 10.0.0.0/8 gw 10.168.128.1" >> $net_config_path
  #echo "up route add -net 172.16.0.0/12 gw 10.168.128.1" >> $net_config_path
  #echo "up route add -net 192.168.0.0/16 gw 10.168.128.1" >> $net_config_path
  bridges=$(sed ":a;N;s/\n/ /g" $device_tmpfile)
  echo ${bridges}
}

function make_hostname() {
  hostname=$1
  sed -i "1s/.*/${hostname}/g"  $TMPDIR/etc/hostname
  sed -i "2s/.*/127.0.0.1\t${hostname}/g"  $TMPDIR/etc/hosts
  #sed -i "s/HOSTN/HOSTNAME=${hostname}/g"  $TMPDIR/etc/sysconfig/network
  #sed -i "/HOSTNAME/ a NOZEROCONF=yes"  $TMPDIR/etc/sysconfig/network
}


make_others() {
#  aptserver=$1
  #sed -i "s/NTPSERVER/$ntpserver/g"  $TMPDIR/etc/ntp.conf
  #sed -i "s/#ADDRESS_APT/$aptserver/g" $TMPDIR/etc/hosts
  #sed -i "s/ADDRESS_YUM/$yumserver/g" $TMPDIR/etc/yum.repos.d/*
  #sed -i "s/YUM_SITE/$yumsite/g" $TMPDIR/etc/yum.repos.d/*
  #sed -i "s/ceph-9.2.0/ceph_10_2_1/g" $TMPDIR/etc/yum.repos.d/ceph.repo
  #ssh_key
  ssh_dir="$TMPDIR/root/.ssh"
  if [ ! -d ${ssh_dir} ];then
    mkdir -p ${ssh_dir}
  fi
  cat << EOF > $ssh_dir/id_rsa
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvKjMRjjs7HyYaFf9TY/sjsnMZdO17DJRcusDWAA4WO6YyBdj
DkastHhTVHJD6j49a39rn8i3SNcUhW8SAhCsMLUJlJzhfxMeGDNvBqgZoGgzIK4R
O+6mzeTpNsEj3JEtlFAZ7Iwxu1dC5sUf2cdDwshSryO7EqzAc4ZslF/yn4PgCXOY
gqah+0nvZZ4xxtzHX5Wxoevo6F+jY90Ld5kAWvJvU68AZN3TgnTMcSUnfEWVLM9D
6lVwqCl/bQ6HS11qZzoGuEO+TKGfn3ajUW0a+7i8l4oDg0Wal0diI92uHhQM7N4T
DtezyCrzHiX2YnOFeQzwJ7MXCQav0aL0bXJJAwIDAQABAoIBABywVo//LBgyQkLr
znsy/bgg+9IoRavrYvNkxZdmQStU7SrQU6HiIXU4LwPSdH10hYaJU+ZycVzESDya
TVS/EFA/13sf+DKIx7TKbYHHok4ASnYXwkso2XjJ7KUE7d0mvpWlMKwGDbH9bREG
vPczFBzUta4octQ+LO3kbTKK/KxA8MGWclf3Dcp4D77BtZwwPfRQTE7eiiUzr1mm
sakHazS3RyBVJ+qKgNTRPQUM9b5OKZWbsO7np+bh76l/TsdnqVsuQomDp3eSXB+u
XMMaEBNUzeDBTnWnV+krte0PO1E6juaPlYgo4N1qBizZsG4/9KIYAWDgblgJISL/
Me48oIECgYEA33PLKZb2hJDVAl60NqeHMuqv1s+E+Z0Bq1KV3ab6HXVXePvgSSIr
9EXbnB+ljQ36qxhjKuJRWdPSlcDj/NiJ5DHszrAy6hwt0VXEoJFeshTV5QQMGIku
8ofwQjGxvJ8HvLxZbpx6wKVYln4c3cyO2dQjXJhzfW2Wg5E+hwV/VuECgYEA2COg
a6xo6nbkqQmMUMRANFICCyKtzkRibKdXovTP5EfSPte5IdHDM49cd9MQ3tRTPWpW
3brI4oQl1gK1JIC8zA5elzwNT2fWtbrPVZR4KPS8ExRPIr7BfQvOIblsnUgHbege
Z4Xf+u/9RS0C2MedL4WOtUpB1EPbnln35nWZsGMCgYBkGQ0TjmrUK8UgEKiOKzHn
XzZx9fhNkUXkJ1S3PEui0qPisIJigIpMHNcp8wtIStDVwFD8LvHeWYNmTkhTRfVP
YgYA+PF16jWkJtW0UCqpf6fptYxtmVaMktTP8k76fgsLQxyU7kgW8Hrkv43S1gXQ
ErXcjvZ9Y5AfU/s8pPvMQQKBgQCTAdBXy+FkL8+gxVTBjmnY7DplloW+uLZ1DnDF
7lsD+nGeup05ynFJPWX4Pf/If4PKTuycTTHbF2SgpiDMnh9Lby6ZEIhBDPB1lIT2
wU/lE2hkVbjpefMieQgP2g1tAJPFBk6/vMe15stN5Kp+BW785otE9SfHFwxmLO02
u5/33wKBgQCVh3uM7lZLkFhoX0xCND3ag4WoLYWIXNwB1SP9+JHTY2zOsbpnjUbp
PYVepGxHSN9Nw+XSwYOgA9+f9J7H+EUnfOjZo3TUHnkb/qbZwgOU7cJuMPfaJ6Qe
NE+ZPPMgoSF56yDyiX3d2tlDI6qfrSSj/4gYy0KQb5ASSmL0DPvzwA==
-----END RSA PRIVATE KEY-----
EOF
  cat << EOF > $ssh_dir/authorized_keys
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDGQCauelksDsNmWw81xsPdx2ptRA37zef3B3nRZGOBQjFV0/fPal7MFaWmk0aZT/i7nhrgFU4WQtJ+viWI4VRejVrfw1oO/6GV7ar9FbK5sPSXcViqtFuTpdvfAncxVGWmVn1yCYNC+8ffFy0+XrPt9REdPwyKBvnUcLbP2ohusKcdPpS5IbdoIZrxMHjQD4RwUatspNyoP3sxFLrqi2Tj0n+hovEjCxRuJGeTzOTarQpx6v8ANqb35WedRkkyPCOAQmAiUIh7JQaqkcGTy0ZRGwemSylIZRbUWyLl/nzZJLRkL539URDry/M5+SbqwdekPi5IIvDK5rW7aIgFSkpj lihg@KAIXIANGdeMacBook-Pro-2.local
EOF
  cat << EOF > $ssh_dir/config
Host *
  IdentityFile ~/.ssh/id_rsa
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  CheckHostIP no
  ServerAliveCountMax 10
  ServerAliveInterval 60
  TCPKeepAlive no
  SendENV LANG
EOF
  chmod -R 600 ${ssh_dir}
  #add
#  cat ${WORKDIR}/hosts >> $TMPDIR/etc/hosts
#  cp ${WORKDIR}/get.sh $TMPDIR/root/
  #echo "bash /root/get.sh" >> $TMPDIR/etc/rc.local
  chmod +x $TMPDIR/etc/rc.local
  echo "ping -c 2 192.168.206.254" >> $TMPDIR/etc/rc.local
}

createvm() {
  hostname=$1
  _ram=$2
  ram=$[_ram*1024]
  vcpus=$3
  bridges=$4
  osds=$5
  virt-install --connect qemu:///system --name ${hostname} --ram $ram \
--vcpus $vcpus --disk $MNTDIR/${hostname}.qcow2,format=qcow2,cache=none \
--nonetworks --noautoconsole --graphics vnc,listen=0.0.0.0 \
--os-type linux --import --hvm --security type=none \
&& virsh dominfo ${hostname} && echo -e "!!!!!!!!--Created vm ${hostname} successfully " \
||echo "!!!!!!!!!!!!!--Create VmOS ${hostname} failed--!!!!!!!!!!!!"
  for br in ${bridges};do
    echo "Attach interface to $hostname"
    virsh attach-interface ${hostname} bridge $br --model virtio --config
  done
  disks=([1]="sdb" [2]="sdc" [3]="sdd")
  if [ $osds -gt 0 ];then
    for((c=1;c<=$osds;c++));do
      osd_disk="$MNTDIR/$hostname-disk-$c.qcow2"
      if [ -s $osd_disk ];then
        rm -f $osd_disk
      fi
      echo "Attach disk to $hostname"
      qemu-img create -f qcow2 $osd_disk 200G
      chmod 777 $osd_disk
      virsh attach-disk $hostname $osd_disk ${disks[$c]} --cache=none --subdriver=qcow2 --config
    done
  fi
  virsh destroy $hostname
  virsh start $hostname
}


run () {
  host=$1
  devs=$2
  vcpu=$3
  mem=$4
  osd=$5
  pre ${host}
  echo "###Start to create vm ${host}"
  \cp ${BASEIMG} ${MNTDIR}/${host}.qcow2
  chmod 777 ${MNTDIR}/${host}.qcow2
  echo -e "Step1: Guestmount ${host}.qcow2"
  guestmount -a ${MNTDIR}/${host}.qcow2 -i ${TMPDIR}
  make_hostname ${host}
  brs=$(make_net_config_and_return_bridge ${devs})
#  make_others $6
  make_others
  umount $TMPDIR
  sync
  sleep 5
  echo -e "Step2: Import ${host}.qcow2"
  createvm ${host} ${mem} ${vcpu} "${brs}" $osd
  echo "###End ${host}"
}


########
#run "test-4" "br-mgmt#192.168.206.114#255.255.255.0#192.168.206.254" 4 4 0
#create_vms.sh {{item.vmname}} "{{BRIDGE}}#{{item.ip}}#{{NET_MASK}}#{{NET_GATEWAY}}#DNS" {{item.vcpu}} {{item.vmem}} {{item.osd}} {{APT_SOURCE}} {{WORKDIR}}/{{IMAGE_FILE_NAME}}
run $1 $2 $3 $4 $5 $6

#run "ssh-download" "br-mgmt#192.168.206.115#255.255.255.0#192.168.206.254" 4 4 0
