# pw

A simple, deterministic password generator based on HMAC-SHA256.

It generates unique, complex passwords for different sites based on a master secret key. This means you only need to remember one secret, and you can recreate the same password for any site whenever you need it.

## How It Works

The core logic is straightforward:

1.  It uses a secret key you provide.
2.  It takes a site name (e.g., `github.com`) as input.
3.  It calculates `HMAC-SHA256(secret_key, site_name)`.
4.  The resulting hash is encoded into a Base62 string.
5.  A site-specific policy is applied to the string to meet requirements like length, uppercase, lowercase, digits, or special characters.

## Setup

1.  Create a configuration directory:

    ```sh
    mkdir -p ~/.config/pw
    ```

2.  Create your secret key file. The key should be a long, high-entropy string.

    ```sh
    # For example, using /dev/urandom
    head -c 64 /dev/urandom | base64 > ~/.config/pw/secret
    ```

3.  Set the correct permissions for the secret file. The program will refuse to run if the permissions are not `0600`.
    ```sh
    chmod 0600 ~/.config/pw/secret
    ```

## Usage

Pipe a site name to the `pw` command:

```sh
pw
# then enter the site name in stdin
```

The generated password for that site will be printed to standard output.

## Site Policies

For sites with specific password requirements (e.g., "must contain a special character"), you can create a policy file.

Create a directory for site policies:

```sh
mkdir -p ~/.config/pw/sites
```

Create a configuration file named after the site. For example, for `google.com`:

```
# File: ~/.config/pw/sites/google.com.conf

len=20
upper=true
lower=true
digit=true
special=true
special_chars=!@#$%^&*
```

### Supported Keys

| Key             | Description                             | Default Value |
| --------------- | --------------------------------------- | ------------- |
| `len`           | Desired password length.                | `12`          |
| `upper`         | Require at least one uppercase letter.  | `false`       |
| `lower`         | Require at least one lowercase letter.  | `false`       |
| `digit`         | Require at least one digit.             | `false`       |
| `special`       | Require at least one special character. | `false`       |
| `special_chars` | A string of special characters to use.  | `!@#$%^&*`    |

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
