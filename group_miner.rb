require "digest"
require "json"
require "faraday"

class GroupMiner
  NUM_PROCS = 4
  BASE_URL = "http://git-coin.herokuapp.com"
  REFRESH_THRESHOLD = 1_000_000

  def mine(i)
    puts "miner #{i} starting to mine!"
    iteration = 0
    current_target = fetch_target
    message = seed_message
    start = Time.now

    loop do
      digest = Digest::SHA1.hexdigest(message)
      if digest < current_target
        response = JSON.parse(submit_hash(message))
        if response["success"]
          puts "got a hash! #{response}"
          current_target = response["new_target"]
          iteration = 0
          start = Time.now
        else
          puts "aw too bad; our guess got beat out :( #{response}"
          current_target = response["new_target"]
        end
      end

      message = digest
      if iteration > REFRESH_THRESHOLD
        puts "completed #{REFRESH_THRESHOLD} in #{Time.now - start} seconds"
        iteration = 0
        puts "last guess was: #{message}; current target: #{current_target}"
        puts "completed #{REFRESH_THRESHOLD} iteration, refreshing target"
        current_target = fetch_target
      end

      iteration += 1
    end
  end

  def seed_message
    rand.to_s
  end

  def submit_hash(message)
    Faraday.post(BASE_URL + "/hash", {:owner => "turing ghost", :message => message}).body
  end

  def fetch_target
    Faraday.get(BASE_URL + "/target").body
  end
end


4.times.map do |i|
  Thread.new do
    GroupMiner.new.mine(i)
  end
end.map(&:join)

