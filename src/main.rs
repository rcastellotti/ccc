use std::collections::HashMap;

#[derive(Default, Debug)]
struct CurlCommand {
    url: Option<String>,
    headers: Vec<String>,
    other_options: Vec<String>,
    method: Option<String>,
}

fn parse_curl_command<I: Iterator<Item = String>>(mut args: I) -> CurlCommand {
    let mut curl_command = CurlCommand::default();
    while let Some(arg) = args.next() {
        if arg == "-H" {
            curl_command
                .headers
                .push(args.next().expect("-H is not followed by an argument"));
        } else if arg == "-X" {
            curl_command.method = Some(args.next().expect("-X is not followed by an argument"));
        } else if arg.starts_with('-') {
            curl_command.other_options.push(arg);
        } else {
            curl_command.url = Some(arg);
        }
    }

    curl_command
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<_> = std::env::args().collect();
    if args.len() < 2 {
        eprintln!("missing arguments");
        return Ok(());
    }

    // skip this program name and curl
    let args = args.into_iter().skip(2);
    let result = parse_curl_command(args);

    dbg!(result);

    Ok(())
}
