fn twice(x: i32) -> i32 {
    x * 2
}

fn main() {
    let mut v: Vec<i32> = Vec::new();
    v.push(1);
    let y: i32 = "text";
    println!("{}", twice(y));
}
