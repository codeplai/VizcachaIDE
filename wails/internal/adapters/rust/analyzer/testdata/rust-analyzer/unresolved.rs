struct Counter {
    total: i32,
}

impl Counter {
    fn add(&mut self, n: i32) {
        self.total += n;
    }
}

fn twice(x: i32) -> i32 {
    x * 2
}

fn main() {
    let mut v: Vec<i32> = Vec::new();
    v.push(1);
    let y = twice(2);
    let s = String::from("hi");
    let t = s;
    println!("{} {} {}", missing, t, y);
}
