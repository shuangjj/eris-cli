#!/bin/bash

echo
echo ">>> Sanity checks"
echo
if [ ! -d ${RELEASE_DIR_MAIN} ] || [ ! -d ${RELEASE_DIR_EXPERIMENTAL} ]
then
    echo "Main release and experimental release directories must be set!"
    exit 1
fi
PKGS_MAIN=$(find ${RELEASE_DIR_MAIN} -name "eris_${ERIS_VERSION}-*.deb" -exec echo {} +)
PKGS_EXPERIMENTAL=$(find ${RELEASE_DIR_EXPERIMENTAL} -name "eris_${ERIS_VERSION}-*.deb" -exec echo {} +)

echo "Releasing ${PKGS_MAIN} as 'main'"
echo "          ${PKGS_EXPERIMENTAL} as 'experimental'"

echo
echo ">>> Importing GPG keys"
echo
gpg --import linux-public-key.asc
gpg --import linux-private-key.asc

echo
echo ">>> Copying Debian packages to Amazon S3"
echo
cat > ${HOME}/.s3cfg <<EOF
[default]
access_key = ${AWS_ACCESS_KEY}
secret_key = ${AWS_SECRET_ACCESS_KEY}
EOF
#s3cmd put ${PACKAGE} s3://${AWS_S3_DEB_PACKAGES}

#if [ "$ERIS_BRANCH" != "master" ] 
#then
   #echo
   #echo ">>> Not recreating a repo for #${ERIS_BRANCH} branch"
   #echo
   #exit 0
#fi

echo
echo ">>> Creating an APT repository"
echo
mkdir -p eris/conf
gpg --armor --export "${KEY_NAME}" > eris/APT-GPG-KEY

cat > eris/conf/options <<EOF
verbose
basedir /root/eris
ask-passphrase
EOF

DISTROS="precise trusty utopic vivid wheezy jessie stretch wily xenial"
for distro in ${DISTROS}
do
  cat >> eris/conf/distributions <<EOF
Origin: Eris Industries <support@erisindustries.com>
Codename: ${distro}
Components: main experimental
Architectures: i386 amd64 armhf
SignWith: $(gpg --keyid-format=long --list-keys --with-colons|fgrep "${KEY_NAME}"|cut -d: -f5)

EOF
done

for distro in ${DISTROS}
do
  echo
  echo ">>> Adding package to ${distro}"
  echo
  expect <<-EOF
    set timeout 5
    spawn reprepro -C main -Vb eris includedeb ${distro} ${PKGS_MAIN}
    expect {
            timeout                    { send_error "Failed to submit password"; exit 1 }
            "Please enter passphrase:" { send -- "${KEY_PASSWORD}\r";
                                         send_user "********";
                                         exp_continue
                                       }
    }
    wait
    exit 0
EOF
  expect <<-EOF
      set timeout 5
      spawn reprepro -C experimental -Vb eris includedeb ${distro} ${PKGS_EXPERIMENTAL}
      expect {
              timeout                    { send_error "Failed to submit password"; exit 1 }
              "Please enter passphrase:" { send -- "${KEY_PASSWORD}\r";
                                           send_user "********";
                                           exp_continue
                                         }
      }
      wait
      exit 0
EOF
done

echo
echo ">>> After adding we have the following"
echo
reprepro -b eris ls eris

echo
echo ">>> Syncing repos to Amazon S3"
echo
s3cmd sync eris/APT-GPG-KEY s3://${AWS_S3_DEB_REPO}
s3cmd sync eris/db s3://${AWS_S3_DEB_REPO}
s3cmd sync eris/dists s3://${AWS_S3_DEB_REPO}
s3cmd sync eris/pool s3://${AWS_S3_DEB_REPO}

echo
echo ">>> Installation instructions"
echo
echo "  \$ curl https://${AWS_S3_DEB_REPO}.s3.amazonaws.com/APT-GPG-KEY | apt-key add -"
echo "  \$ echo \"deb https://${AWS_S3_DEB_REPO}.s3.amazonaws.com DIST main\" | sudo tee /etc/apt/sources.list.d/eris.list"
echo
echo "  If you're a fearless marmot who wants new features, add the 'experimental' repo"
echo "  \$ echo \"deb https://${AWS_S3_DEB_REPO}.s3.amazonaws.com DIST experimental\" | sudo tee -a /etc/apt/sources.list.d/eris.list"
echo
echo "  \$ apt-get update"
echo "  \$ apt-get install eris"
echo
