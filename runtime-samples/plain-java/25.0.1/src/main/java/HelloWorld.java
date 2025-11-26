public class HelloWorld {

    public static void main(String[] args){
        while (!Thread.currentThread().isInterrupted()) {
            System.out.println("Hello World");
            try {
                Thread.sleep(1000);
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }
        }
    }
}
