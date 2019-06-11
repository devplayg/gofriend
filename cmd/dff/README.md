# Duplicate File Finder

중복 파일 검색

### 특정 디렉토리에서 중복 파일 찾기 

    ./dff -d /dir
    
### 멀티 디렉토리에서 중복 파일 찾기

    ./dff -d /dir1 -d /dir2 -d /dir3
    
### 특정 디렉토리에서, 최소 10개 이상 중복되는 파일 찾기 (기본: 5)

    ./dff -d /dir -c 10
    
### 특정 디렉토리에서 파일크기가 2MB 이상인 파일 중에 중복되는 파일 찾기 (기본: 1,000,000 Bytes)

    ./dff -d /dir -s 2000000 

### 결과 정렬

#### 중복되는 파일 총 용량 기준 정렬 (기본)

    ./dff -d /dir -r total

#### 단위파일 크기 기준 정렬
 
    ./dff -d /dir -r size
    
#### 중복되는 파일 개수 기준 정렬    
    
    ./dff -d /dir -r count
    